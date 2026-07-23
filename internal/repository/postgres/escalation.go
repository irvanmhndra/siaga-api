package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type escalationRepo struct{ db *sqlx.DB }

func NewEscalationPolicyRepository(db *sqlx.DB) *escalationRepo { return &escalationRepo{db} }

func (r *escalationRepo) Create(ctx context.Context, p *model.EscalationPolicy) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO escalation_policies (id, organization_id, name, repeat_count) VALUES ($1, $2, $3, $4)`,
		p.ID, p.OrganizationID, p.Name, p.RepeatCount,
	)
	return err
}

func (r *escalationRepo) FindByID(ctx context.Context, id, orgID string) (*model.EscalationPolicy, error) {
	var p model.EscalationPolicy
	err := r.db.GetContext(ctx, &p,
		`SELECT * FROM escalation_policies WHERE id = $1 AND organization_id = $2`, id, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if p.Rules, err = r.ListRules(ctx, p.ID); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *escalationRepo) ListByOrg(ctx context.Context, orgID string) ([]model.EscalationPolicy, error) {
	policies := []model.EscalationPolicy{}
	err := r.db.SelectContext(ctx, &policies,
		`SELECT * FROM escalation_policies WHERE organization_id = $1 ORDER BY name`, orgID)
	if err != nil {
		return nil, err
	}
	for i := range policies {
		if policies[i].Rules, err = r.ListRules(ctx, policies[i].ID); err != nil {
			return nil, err
		}
	}
	return policies, nil
}

func (r *escalationRepo) Update(ctx context.Context, p *model.EscalationPolicy) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE escalation_policies SET name = $1, repeat_count = $2, updated_at = NOW()
		 WHERE id = $3 AND organization_id = $4`,
		p.Name, p.RepeatCount, p.ID, p.OrganizationID,
	)
	return err
}

func (r *escalationRepo) Delete(ctx context.Context, id, orgID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM escalation_policies WHERE id = $1 AND organization_id = $2`, id, orgID)
	return err
}

func (r *escalationRepo) ReplaceRules(ctx context.Context, policyID string, rules []model.EscalationRule) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Deleting the rules cascades to their targets (FK ON DELETE CASCADE).
	if _, err := tx.ExecContext(ctx, `DELETE FROM escalation_rules WHERE escalation_policy_id = $1`, policyID); err != nil {
		return err
	}
	for i, rule := range rules {
		var ruleID string
		if err := tx.GetContext(ctx, &ruleID,
			`INSERT INTO escalation_rules (id, escalation_policy_id, position, delay_minutes)
			 VALUES (gen_random_uuid(), $1, $2, $3) RETURNING id`,
			policyID, i, rule.DelayMinutes,
		); err != nil {
			return err
		}
		for j, t := range rule.Targets {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO escalation_targets (id, escalation_rule_id, target_type, target_id, position)
				 VALUES (gen_random_uuid(), $1, $2, $3, $4)`,
				ruleID, t.TargetType, t.TargetID, j,
			); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (r *escalationRepo) ListRules(ctx context.Context, policyID string) ([]model.EscalationRule, error) {
	rules := []model.EscalationRule{}
	err := r.db.SelectContext(ctx, &rules,
		`SELECT * FROM escalation_rules WHERE escalation_policy_id = $1 ORDER BY position`, policyID)
	if err != nil {
		return nil, err
	}
	for i := range rules {
		targets := []model.EscalationTarget{}
		if err := r.db.SelectContext(ctx, &targets,
			`SELECT * FROM escalation_targets WHERE escalation_rule_id = $1 ORDER BY position`, rules[i].ID); err != nil {
			return nil, err
		}
		rules[i].Targets = targets
	}
	return rules, nil
}
