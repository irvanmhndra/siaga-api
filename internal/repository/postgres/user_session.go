package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type userSessionRepo struct{ db *sqlx.DB }

func NewUserSessionRepository(db *sqlx.DB) *userSessionRepo { return &userSessionRepo{db} }

func (r *userSessionRepo) Create(ctx context.Context, s *model.UserSession) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_sessions (id, user_id, token, refresh_token, expires_at, refresh_expires_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		s.ID, s.UserID, s.Token, s.RefreshToken, s.ExpiresAt, s.RefreshExpiresAt,
	)
	return err
}

func (r *userSessionRepo) FindByToken(ctx context.Context, token string) (*model.UserSession, error) {
	var s model.UserSession
	err := r.db.GetContext(ctx, &s, `SELECT * FROM user_sessions WHERE token = $1`, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *userSessionRepo) FindByRefreshToken(ctx context.Context, refreshToken string) (*model.UserSession, error) {
	var s model.UserSession
	err := r.db.GetContext(ctx, &s, `SELECT * FROM user_sessions WHERE refresh_token = $1`, refreshToken)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *userSessionRepo) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE token = $1`, token)
	return err
}

func (r *userSessionRepo) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE user_id = $1`, userID)
	return err
}
