package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/irvanmhndra/siaga-api/internal/repository"
	"github.com/irvanmhndra/siaga-api/pkg/apperror"
)

type EscalationService struct {
	policies repository.EscalationPolicyRepository
}

func NewEscalationService(policies repository.EscalationPolicyRepository) *EscalationService {
	return &EscalationService{policies}
}

type TargetRequest struct {
	TargetType string `json:"target_type" validate:"required,oneof=user schedule"`
	TargetID   string `json:"target_id" validate:"required"`
}

type RuleRequest struct {
	DelayMinutes int             `json:"delay_minutes" validate:"min=0"`
	Targets      []TargetRequest `json:"targets" validate:"required,min=1,dive"`
}

type PolicyRequest struct {
	Name        string        `json:"name" validate:"required,min=2,max=100"`
	RepeatCount int           `json:"repeat_count" validate:"min=0,max=10"`
	Rules       []RuleRequest `json:"rules" validate:"required,min=1,dive"`
}

func (s *EscalationService) Create(ctx context.Context, orgID string, req PolicyRequest) (*model.EscalationPolicy, error) {
	p := &model.EscalationPolicy{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		Name:           req.Name,
		RepeatCount:    req.RepeatCount,
	}
	if err := s.policies.Create(ctx, p); err != nil {
		return nil, err
	}
	if err := s.policies.ReplaceRules(ctx, p.ID, toRules(req.Rules)); err != nil {
		return nil, err
	}
	return s.Get(ctx, p.ID, orgID)
}

func (s *EscalationService) List(ctx context.Context, orgID string) ([]model.EscalationPolicy, error) {
	return s.policies.ListByOrg(ctx, orgID)
}

func (s *EscalationService) Get(ctx context.Context, id, orgID string) (*model.EscalationPolicy, error) {
	p, err := s.policies.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, apperror.NotFound("escalation policy not found")
	}
	return p, nil
}

func (s *EscalationService) Update(ctx context.Context, id, orgID string, req PolicyRequest) (*model.EscalationPolicy, error) {
	p, err := s.Get(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	p.Name = req.Name
	p.RepeatCount = req.RepeatCount
	if err := s.policies.Update(ctx, p); err != nil {
		return nil, err
	}
	if req.Rules != nil {
		if err := s.policies.ReplaceRules(ctx, p.ID, toRules(req.Rules)); err != nil {
			return nil, err
		}
	}
	return s.Get(ctx, id, orgID)
}

func (s *EscalationService) Delete(ctx context.Context, id, orgID string) error {
	if _, err := s.Get(ctx, id, orgID); err != nil {
		return err
	}
	return s.policies.Delete(ctx, id, orgID)
}

func toRules(reqs []RuleRequest) []model.EscalationRule {
	rules := make([]model.EscalationRule, 0, len(reqs))
	for _, r := range reqs {
		targets := make([]model.EscalationTarget, 0, len(r.Targets))
		for _, t := range r.Targets {
			targets = append(targets, model.EscalationTarget{TargetType: t.TargetType, TargetID: t.TargetID})
		}
		rules = append(rules, model.EscalationRule{DelayMinutes: r.DelayMinutes, Targets: targets})
	}
	return rules
}
