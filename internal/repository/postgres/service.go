package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type serviceRepo struct{ db *sqlx.DB }

func NewServiceRepository(db *sqlx.DB) *serviceRepo { return &serviceRepo{db} }

func (r *serviceRepo) Create(ctx context.Context, s *model.Service) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO services
		 (id, organization_id, name, description, escalation_policy_id, integration_key, status,
		  auto_resolve_minutes, ack_timeout_minutes, slack_webhook_url)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		s.ID, s.OrganizationID, s.Name, s.Description, s.EscalationPolicyID, s.IntegrationKey, s.Status,
		s.AutoResolveMinutes, s.AckTimeoutMinutes, s.SlackWebhookURL,
	)
	return err
}

func (r *serviceRepo) FindByID(ctx context.Context, id, orgID string) (*model.Service, error) {
	var s model.Service
	err := r.db.GetContext(ctx, &s,
		`SELECT * FROM services WHERE id = $1 AND organization_id = $2`, id, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *serviceRepo) FindByIntegrationKey(ctx context.Context, key string) (*model.Service, error) {
	var s model.Service
	err := r.db.GetContext(ctx, &s,
		`SELECT * FROM services WHERE integration_key = $1`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *serviceRepo) ListByOrg(ctx context.Context, orgID string) ([]model.Service, error) {
	services := []model.Service{}
	err := r.db.SelectContext(ctx, &services,
		`SELECT * FROM services WHERE organization_id = $1 ORDER BY name`, orgID)
	return services, err
}

func (r *serviceRepo) Update(ctx context.Context, s *model.Service) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE services SET name = $1, description = $2, escalation_policy_id = $3, status = $4,
		 auto_resolve_minutes = $5, ack_timeout_minutes = $6, slack_webhook_url = $7, updated_at = NOW()
		 WHERE id = $8 AND organization_id = $9`,
		s.Name, s.Description, s.EscalationPolicyID, s.Status,
		s.AutoResolveMinutes, s.AckTimeoutMinutes, s.SlackWebhookURL, s.ID, s.OrganizationID,
	)
	return err
}

func (r *serviceRepo) Delete(ctx context.Context, id, orgID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM services WHERE id = $1 AND organization_id = $2`, id, orgID)
	return err
}
