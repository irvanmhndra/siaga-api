package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/irvanmhndra/siaga-api/internal/repository"
	"github.com/irvanmhndra/siaga-api/pkg/apperror"
)

// ServiceService manages monitored services — the things that page. Each owns a
// unique integration_key used by the Events API to route events in.
type ServiceService struct {
	services   repository.ServiceRepository
	policies   repository.EscalationPolicyRepository
}

func NewServiceService(services repository.ServiceRepository, policies repository.EscalationPolicyRepository) *ServiceService {
	return &ServiceService{services, policies}
}

type ServiceRequest struct {
	Name               string  `json:"name" validate:"required,min=2,max=120"`
	Description        *string `json:"description"`
	EscalationPolicyID *string `json:"escalation_policy_id"`
	Status             string  `json:"status" validate:"omitempty,oneof=active maintenance disabled"`
	AutoResolveMinutes *int    `json:"auto_resolve_minutes" validate:"omitempty,min=1"`
	AckTimeoutMinutes  *int    `json:"ack_timeout_minutes" validate:"omitempty,min=1"`
	SlackWebhookURL    *string `json:"slack_webhook_url"`
}

func (s *ServiceService) Create(ctx context.Context, orgID string, req ServiceRequest) (*model.Service, error) {
	if err := s.validatePolicy(ctx, orgID, req.EscalationPolicyID); err != nil {
		return nil, err
	}
	key, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	svc := &model.Service{
		ID:                 uuid.NewString(),
		OrganizationID:     orgID,
		Name:               req.Name,
		Description:        req.Description,
		EscalationPolicyID: req.EscalationPolicyID,
		IntegrationKey:     key,
		Status:             status,
		AutoResolveMinutes: req.AutoResolveMinutes,
		AckTimeoutMinutes:  req.AckTimeoutMinutes,
		SlackWebhookURL:    req.SlackWebhookURL,
	}
	if err := s.services.Create(ctx, svc); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *ServiceService) List(ctx context.Context, orgID string) ([]model.Service, error) {
	return s.services.ListByOrg(ctx, orgID)
}

func (s *ServiceService) Get(ctx context.Context, id, orgID string) (*model.Service, error) {
	svc, err := s.services.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, apperror.NotFound("service not found")
	}
	return svc, nil
}

func (s *ServiceService) Update(ctx context.Context, id, orgID string, req ServiceRequest) (*model.Service, error) {
	svc, err := s.Get(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if err := s.validatePolicy(ctx, orgID, req.EscalationPolicyID); err != nil {
		return nil, err
	}
	svc.Name = req.Name
	svc.Description = req.Description
	svc.EscalationPolicyID = req.EscalationPolicyID
	if req.Status != "" {
		svc.Status = req.Status
	}
	svc.AutoResolveMinutes = req.AutoResolveMinutes
	svc.AckTimeoutMinutes = req.AckTimeoutMinutes
	svc.SlackWebhookURL = req.SlackWebhookURL
	if err := s.services.Update(ctx, svc); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *ServiceService) Delete(ctx context.Context, id, orgID string) error {
	if _, err := s.Get(ctx, id, orgID); err != nil {
		return err
	}
	return s.services.Delete(ctx, id, orgID)
}

func (s *ServiceService) validatePolicy(ctx context.Context, orgID string, policyID *string) error {
	if policyID == nil || *policyID == "" {
		return nil
	}
	p, err := s.policies.FindByID(ctx, *policyID, orgID)
	if err != nil {
		return err
	}
	if p == nil {
		return apperror.BadRequest("invalid escalation policy")
	}
	return nil
}
