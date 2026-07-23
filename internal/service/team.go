package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/irvanmhndra/siaga-api/internal/repository"
	"github.com/irvanmhndra/siaga-api/pkg/apperror"
)

type TeamService struct {
	teams repository.TeamRepository
}

func NewTeamService(teams repository.TeamRepository) *TeamService {
	return &TeamService{teams}
}

type TeamRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

func (s *TeamService) Create(ctx context.Context, orgID string, req TeamRequest) (*model.Team, error) {
	t := &model.Team{ID: uuid.NewString(), OrganizationID: orgID, Name: req.Name}
	if err := s.teams.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TeamService) List(ctx context.Context, orgID string) ([]model.Team, error) {
	teams, err := s.teams.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	for i := range teams {
		members, err := s.teams.ListMembers(ctx, teams[i].ID)
		if err != nil {
			return nil, err
		}
		teams[i].Members = members
	}
	return teams, nil
}

func (s *TeamService) Get(ctx context.Context, id, orgID string) (*model.Team, error) {
	t, err := s.teams.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, apperror.NotFound("team not found")
	}
	if t.Members, err = s.teams.ListMembers(ctx, t.ID); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TeamService) Update(ctx context.Context, id, orgID string, req TeamRequest) (*model.Team, error) {
	t, err := s.Get(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	t.Name = req.Name
	if err := s.teams.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TeamService) Delete(ctx context.Context, id, orgID string) error {
	if _, err := s.Get(ctx, id, orgID); err != nil {
		return err
	}
	return s.teams.Delete(ctx, id, orgID)
}

func (s *TeamService) AddMember(ctx context.Context, teamID, orgID, userID string) error {
	if _, err := s.Get(ctx, teamID, orgID); err != nil {
		return err
	}
	return s.teams.AddMember(ctx, teamID, userID)
}

func (s *TeamService) RemoveMember(ctx context.Context, teamID, orgID, userID string) error {
	if _, err := s.Get(ctx, teamID, orgID); err != nil {
		return err
	}
	return s.teams.RemoveMember(ctx, teamID, userID)
}
