package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type userRepo struct{ db *sqlx.DB }

func NewUserRepository(db *sqlx.DB) *userRepo { return &userRepo{db} }

func (r *userRepo) Create(ctx context.Context, u *model.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, organization_id, name, email, password_hash, phone, timezone, notify_email, notify_sms)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		u.ID, u.OrganizationID, u.Name, u.Email, u.PasswordHash, u.Phone, u.Timezone, u.NotifyEmail, u.NotifySMS,
	)
	return err
}

func (r *userRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

func (r *userRepo) ListByOrg(ctx context.Context, orgID string) ([]model.User, error) {
	users := []model.User{}
	err := r.db.SelectContext(ctx, &users,
		`SELECT * FROM users WHERE organization_id = $1 ORDER BY name`, orgID)
	return users, err
}

func (r *userRepo) Update(ctx context.Context, u *model.User) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET name = $1, phone = $2, timezone = $3, notify_email = $4, notify_sms = $5, updated_at = NOW()
		 WHERE id = $6`,
		u.Name, u.Phone, u.Timezone, u.NotifyEmail, u.NotifySMS, u.ID,
	)
	return err
}
