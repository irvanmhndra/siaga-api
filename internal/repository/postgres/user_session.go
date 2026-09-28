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
		`INSERT INTO user_sessions (id, user_id, token_hash, refresh_token_hash, expires_at, refresh_expires_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		s.ID, s.UserID, s.TokenHash, s.RefreshTokenHash, s.ExpiresAt, s.RefreshExpiresAt,
	)
	return err
}

func (r *userSessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*model.UserSession, error) {
	var s model.UserSession
	err := r.db.GetContext(ctx, &s, `SELECT * FROM user_sessions WHERE token_hash = $1`, tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *userSessionRepo) FindByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*model.UserSession, error) {
	var s model.UserSession
	err := r.db.GetContext(ctx, &s, `SELECT * FROM user_sessions WHERE refresh_token_hash = $1`, refreshTokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *userSessionRepo) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE token_hash = $1`, tokenHash)
	return err
}

func (r *userSessionRepo) Rotate(ctx context.Context, oldID string, next *model.UserSession) (bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	// The delete is the claim: of two concurrent refreshes with the same
	// token, only one sees a row affected.
	res, err := tx.ExecContext(ctx, `DELETE FROM user_sessions WHERE id = $1`, oldID)
	if err != nil {
		return false, err
	}
	if n, err := res.RowsAffected(); err != nil || n == 0 {
		return false, err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO user_sessions (id, user_id, token_hash, refresh_token_hash, expires_at, refresh_expires_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		next.ID, next.UserID, next.TokenHash, next.RefreshTokenHash, next.ExpiresAt, next.RefreshExpiresAt,
	)
	if err != nil {
		return false, err
	}
	return true, tx.Commit()
}

func (r *userSessionRepo) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE user_id = $1`, userID)
	return err
}
