package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type scheduleRepo struct{ db *sqlx.DB }

func NewScheduleRepository(db *sqlx.DB) *scheduleRepo { return &scheduleRepo{db} }

func (r *scheduleRepo) Create(ctx context.Context, s *model.Schedule) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO schedules (id, organization_id, name, timezone) VALUES ($1, $2, $3, $4)`,
		s.ID, s.OrganizationID, s.Name, s.Timezone,
	)
	return err
}

func (r *scheduleRepo) FindByID(ctx context.Context, id, orgID string) (*model.Schedule, error) {
	var s model.Schedule
	err := r.db.GetContext(ctx, &s,
		`SELECT * FROM schedules WHERE id = $1 AND organization_id = $2`, id, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if s.Layers, err = r.ListLayers(ctx, s.ID); err != nil {
		return nil, err
	}
	if s.Overrides, err = r.ListOverrides(ctx, s.ID); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *scheduleRepo) ListByOrg(ctx context.Context, orgID string) ([]model.Schedule, error) {
	schedules := []model.Schedule{}
	err := r.db.SelectContext(ctx, &schedules,
		`SELECT * FROM schedules WHERE organization_id = $1 ORDER BY name`, orgID)
	if err != nil {
		return nil, err
	}
	for i := range schedules {
		if schedules[i].Layers, err = r.ListLayers(ctx, schedules[i].ID); err != nil {
			return nil, err
		}
	}
	return schedules, nil
}

func (r *scheduleRepo) Update(ctx context.Context, s *model.Schedule) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE schedules SET name = $1, timezone = $2, updated_at = NOW()
		 WHERE id = $3 AND organization_id = $4`,
		s.Name, s.Timezone, s.ID, s.OrganizationID,
	)
	return err
}

func (r *scheduleRepo) Delete(ctx context.Context, id, orgID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM schedules WHERE id = $1 AND organization_id = $2`, id, orgID)
	return err
}

func (r *scheduleRepo) ReplaceLayers(ctx context.Context, scheduleID string, layers []model.ScheduleLayer) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM schedule_layers WHERE schedule_id = $1`, scheduleID); err != nil {
		return err
	}
	for i, l := range layers {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schedule_layers
			 (id, schedule_id, name, rotation_type, rotation_length_days, handoff_time, start_at, user_ids, position)
			 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8)`,
			scheduleID, l.Name, l.RotationType, l.RotationLengthDays, l.HandoffTime, l.StartAt, l.UserIDs, i,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *scheduleRepo) ListLayers(ctx context.Context, scheduleID string) ([]model.ScheduleLayer, error) {
	layers := []model.ScheduleLayer{}
	err := r.db.SelectContext(ctx, &layers,
		`SELECT * FROM schedule_layers WHERE schedule_id = $1 ORDER BY position`, scheduleID)
	return layers, err
}

func (r *scheduleRepo) CreateOverride(ctx context.Context, o *model.ScheduleOverride) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO schedule_overrides (id, schedule_id, user_id, start_at, end_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		o.ID, o.ScheduleID, o.UserID, o.StartAt, o.EndAt,
	)
	return err
}

func (r *scheduleRepo) DeleteOverride(ctx context.Context, id, scheduleID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM schedule_overrides WHERE id = $1 AND schedule_id = $2`, id, scheduleID)
	return err
}

func (r *scheduleRepo) ListOverrides(ctx context.Context, scheduleID string) ([]model.ScheduleOverride, error) {
	overrides := []model.ScheduleOverride{}
	err := r.db.SelectContext(ctx, &overrides,
		`SELECT * FROM schedule_overrides WHERE schedule_id = $1 ORDER BY start_at`, scheduleID)
	return overrides, err
}
