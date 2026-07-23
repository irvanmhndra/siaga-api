package repository

import (
	"context"
	"time"

	"github.com/irvanmhndra/siaga-api/internal/model"
)

type OrganizationRepository interface {
	Create(ctx context.Context, o *model.Organization) error
	FindByID(ctx context.Context, id string) (*model.Organization, error)
}

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	ListByOrg(ctx context.Context, orgID string) ([]model.User, error)
	Update(ctx context.Context, u *model.User) error
}

type UserSessionRepository interface {
	Create(ctx context.Context, s *model.UserSession) error
	FindByToken(ctx context.Context, token string) (*model.UserSession, error)
	FindByRefreshToken(ctx context.Context, refreshToken string) (*model.UserSession, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID string) error
}

type TeamRepository interface {
	Create(ctx context.Context, t *model.Team) error
	FindByID(ctx context.Context, id, orgID string) (*model.Team, error)
	ListByOrg(ctx context.Context, orgID string) ([]model.Team, error)
	Update(ctx context.Context, t *model.Team) error
	Delete(ctx context.Context, id, orgID string) error
	AddMember(ctx context.Context, teamID, userID string) error
	RemoveMember(ctx context.Context, teamID, userID string) error
	ListMembers(ctx context.Context, teamID string) ([]model.User, error)
}

type ScheduleRepository interface {
	Create(ctx context.Context, s *model.Schedule) error
	FindByID(ctx context.Context, id, orgID string) (*model.Schedule, error)
	ListByOrg(ctx context.Context, orgID string) ([]model.Schedule, error)
	Update(ctx context.Context, s *model.Schedule) error
	Delete(ctx context.Context, id, orgID string) error
	// ReplaceLayers atomically swaps out a schedule's layers.
	ReplaceLayers(ctx context.Context, scheduleID string, layers []model.ScheduleLayer) error
	ListLayers(ctx context.Context, scheduleID string) ([]model.ScheduleLayer, error)
	CreateOverride(ctx context.Context, o *model.ScheduleOverride) error
	DeleteOverride(ctx context.Context, id, scheduleID string) error
	ListOverrides(ctx context.Context, scheduleID string) ([]model.ScheduleOverride, error)
}

type EscalationPolicyRepository interface {
	Create(ctx context.Context, p *model.EscalationPolicy) error
	FindByID(ctx context.Context, id, orgID string) (*model.EscalationPolicy, error)
	ListByOrg(ctx context.Context, orgID string) ([]model.EscalationPolicy, error)
	Update(ctx context.Context, p *model.EscalationPolicy) error
	Delete(ctx context.Context, id, orgID string) error
	// ReplaceRules atomically swaps out a policy's rules and their targets.
	ReplaceRules(ctx context.Context, policyID string, rules []model.EscalationRule) error
	ListRules(ctx context.Context, policyID string) ([]model.EscalationRule, error)
}

type ServiceRepository interface {
	Create(ctx context.Context, s *model.Service) error
	FindByID(ctx context.Context, id, orgID string) (*model.Service, error)
	FindByIntegrationKey(ctx context.Context, key string) (*model.Service, error)
	ListByOrg(ctx context.Context, orgID string) ([]model.Service, error)
	Update(ctx context.Context, s *model.Service) error
	Delete(ctx context.Context, id, orgID string) error
}

// IncidentFilter narrows an incident list query. Empty fields are ignored.
type IncidentFilter struct {
	Status    string
	ServiceID string
}

type IncidentRepository interface {
	Create(ctx context.Context, i *model.Incident) error
	FindByID(ctx context.Context, id, orgID string) (*model.Incident, error)
	FindOpenByDedup(ctx context.Context, serviceID, dedupKey string) (*model.Incident, error)
	ListByOrg(ctx context.Context, orgID string, f IncidentFilter) ([]model.Incident, error)
	Update(ctx context.Context, i *model.Incident) error
	AppendLog(ctx context.Context, e *model.IncidentLogEntry) error
	ListLog(ctx context.Context, incidentID string) ([]model.IncidentLogEntry, error)
	// ListDueForEscalation returns triggered incidents whose next_escalation_at
	// has passed — the escalation worker's work queue.
	ListDueForEscalation(ctx context.Context, now time.Time) ([]model.Incident, error)
	// ListDueForAutoResolve returns open incidents past their service's
	// auto_resolve window.
	ListDueForAutoResolve(ctx context.Context, now time.Time) ([]model.Incident, error)
}

type NotificationRepository interface {
	Create(ctx context.Context, n *model.Notification) error
	ListByUser(ctx context.Context, userID string, unreadOnly bool, limit int) ([]model.Notification, error)
	CountUnread(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, id, userID string) error
	MarkAllRead(ctx context.Context, userID string) error
}
