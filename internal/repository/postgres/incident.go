package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/irvanmhndra/siaga-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type incidentRepo struct{ db *sqlx.DB }

func NewIncidentRepository(db *sqlx.DB) *incidentRepo { return &incidentRepo{db} }

func (r *incidentRepo) Create(ctx context.Context, i *model.Incident) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO incidents
		 (id, organization_id, service_id, title, status, urgency, dedup_key, details,
		  assigned_user_id, escalation_policy_id, escalation_level, escalation_repeats,
		  next_escalation_at, triggered_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		i.ID, i.OrganizationID, i.ServiceID, i.Title, i.Status, i.Urgency, i.DedupKey, i.Details,
		i.AssignedUserID, i.EscalationPolicyID, i.EscalationLevel, i.EscalationRepeats,
		i.NextEscalationAt, i.TriggeredAt,
	)
	return err
}

func (r *incidentRepo) FindByID(ctx context.Context, id, orgID string) (*model.Incident, error) {
	var i model.Incident
	err := r.db.GetContext(ctx, &i,
		`SELECT * FROM incidents WHERE id = $1 AND organization_id = $2`, id, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &i, err
}

func (r *incidentRepo) FindOpenByDedup(ctx context.Context, serviceID, dedupKey string) (*model.Incident, error) {
	var i model.Incident
	err := r.db.GetContext(ctx, &i,
		`SELECT * FROM incidents
		 WHERE service_id = $1 AND dedup_key = $2 AND status <> 'resolved'
		 ORDER BY triggered_at DESC LIMIT 1`, serviceID, dedupKey)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &i, err
}

func (r *incidentRepo) ListByOrg(ctx context.Context, orgID string, f repository.IncidentFilter) ([]model.Incident, error) {
	incidents := []model.Incident{}
	q := `SELECT * FROM incidents WHERE organization_id = $1`
	args := []any{orgID}
	if f.Status != "" {
		args = append(args, f.Status)
		q += " AND status = $" + itoa(len(args))
	}
	if f.ServiceID != "" {
		args = append(args, f.ServiceID)
		q += " AND service_id = $" + itoa(len(args))
	}
	q += " ORDER BY triggered_at DESC LIMIT 200"
	err := r.db.SelectContext(ctx, &incidents, q, args...)
	return incidents, err
}

func (r *incidentRepo) Update(ctx context.Context, i *model.Incident) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE incidents SET status = $1, urgency = $2, assigned_user_id = $3,
		 escalation_level = $4, escalation_repeats = $5, next_escalation_at = $6,
		 acknowledged_at = $7, resolved_at = $8, updated_at = NOW()
		 WHERE id = $9`,
		i.Status, i.Urgency, i.AssignedUserID,
		i.EscalationLevel, i.EscalationRepeats, i.NextEscalationAt,
		i.AcknowledgedAt, i.ResolvedAt, i.ID,
	)
	return err
}

func (r *incidentRepo) AppendLog(ctx context.Context, e *model.IncidentLogEntry) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO incident_log (id, incident_id, type, message, user_id)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4)`,
		e.IncidentID, e.Type, e.Message, e.UserID,
	)
	return err
}

func (r *incidentRepo) ListLog(ctx context.Context, incidentID string) ([]model.IncidentLogEntry, error) {
	log := []model.IncidentLogEntry{}
	err := r.db.SelectContext(ctx, &log,
		`SELECT * FROM incident_log WHERE incident_id = $1 ORDER BY created_at`, incidentID)
	return log, err
}

func (r *incidentRepo) ListDueForEscalation(ctx context.Context, now time.Time) ([]model.Incident, error) {
	incidents := []model.Incident{}
	// 'acknowledged' incidents with a due next_escalation_at are ack-timeouts.
	err := r.db.SelectContext(ctx, &incidents,
		`SELECT * FROM incidents
		 WHERE status IN ('triggered', 'acknowledged')
		   AND next_escalation_at IS NOT NULL AND next_escalation_at <= $1
		 ORDER BY next_escalation_at`, now)
	return incidents, err
}

func (r *incidentRepo) ListDueForAutoResolve(ctx context.Context, now time.Time) ([]model.Incident, error) {
	incidents := []model.Incident{}
	err := r.db.SelectContext(ctx, &incidents,
		`SELECT i.* FROM incidents i
		 JOIN services s ON s.id = i.service_id
		 WHERE i.status <> 'resolved'
		   AND s.auto_resolve_minutes IS NOT NULL
		   AND i.triggered_at + (s.auto_resolve_minutes * interval '1 minute') <= $1`, now)
	return incidents, err
}

// itoa is a tiny helper to build positional placeholders without strconv churn.
func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
