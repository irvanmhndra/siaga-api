package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type orgRepo struct{ db *sqlx.DB }

func NewOrganizationRepository(db *sqlx.DB) *orgRepo { return &orgRepo{db} }

func (r *orgRepo) Create(ctx context.Context, o *model.Organization) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO organizations (id, name) VALUES ($1, $2)`,
		o.ID, o.Name,
	)
	return err
}

func (r *orgRepo) FindByID(ctx context.Context, id string) (*model.Organization, error) {
	var o model.Organization
	err := r.db.GetContext(ctx, &o, `SELECT * FROM organizations WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &o, err
}
