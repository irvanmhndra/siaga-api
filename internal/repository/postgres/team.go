package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type teamRepo struct{ db *sqlx.DB }

func NewTeamRepository(db *sqlx.DB) *teamRepo { return &teamRepo{db} }

func (r *teamRepo) Create(ctx context.Context, t *model.Team) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO teams (id, organization_id, name) VALUES ($1, $2, $3)`,
		t.ID, t.OrganizationID, t.Name,
	)
	return err
}

func (r *teamRepo) FindByID(ctx context.Context, id, orgID string) (*model.Team, error) {
	var t model.Team
	err := r.db.GetContext(ctx, &t,
		`SELECT * FROM teams WHERE id = $1 AND organization_id = $2`, id, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &t, err
}

func (r *teamRepo) ListByOrg(ctx context.Context, orgID string) ([]model.Team, error) {
	teams := []model.Team{}
	err := r.db.SelectContext(ctx, &teams,
		`SELECT * FROM teams WHERE organization_id = $1 ORDER BY name`, orgID)
	return teams, err
}

func (r *teamRepo) Update(ctx context.Context, t *model.Team) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE teams SET name = $1, updated_at = NOW() WHERE id = $2 AND organization_id = $3`,
		t.Name, t.ID, t.OrganizationID,
	)
	return err
}

func (r *teamRepo) Delete(ctx context.Context, id, orgID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM teams WHERE id = $1 AND organization_id = $2`, id, orgID)
	return err
}

func (r *teamRepo) AddMember(ctx context.Context, teamID, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO team_members (team_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		teamID, userID,
	)
	return err
}

func (r *teamRepo) RemoveMember(ctx context.Context, teamID, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, userID)
	return err
}

func (r *teamRepo) ListMembers(ctx context.Context, teamID string) ([]model.User, error) {
	users := []model.User{}
	err := r.db.SelectContext(ctx, &users,
		`SELECT u.* FROM users u
		 JOIN team_members tm ON tm.user_id = u.id
		 WHERE tm.team_id = $1 ORDER BY u.name`, teamID)
	return users, err
}
