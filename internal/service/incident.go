package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/irvanmhndra/siaga-api/internal/repository"
	"github.com/irvanmhndra/siaga-api/pkg/apperror"
)

type IncidentService struct {
	incidents repository.IncidentRepository
	services  repository.ServiceRepository
	policies  repository.EscalationPolicyRepository
	schedules repository.ScheduleRepository
	users     repository.UserRepository
	notifier  *Notifier
}

func NewIncidentService(
	incidents repository.IncidentRepository,
	services repository.ServiceRepository,
	policies repository.EscalationPolicyRepository,
	schedules repository.ScheduleRepository,
	users repository.UserRepository,
	notifier *Notifier,
) *IncidentService {
	return &IncidentService{incidents, services, policies, schedules, users, notifier}
}

// EventRequest is the Events API payload — modeled on PagerDuty's v2 event.
type EventRequest struct {
	RoutingKey  string `json:"routing_key" validate:"required"`
	EventAction string `json:"event_action" validate:"required,oneof=trigger acknowledge resolve"`
	DedupKey    string `json:"dedup_key"`
	Payload     struct {
		Summary  string          `json:"summary"`
		Severity string          `json:"severity"`
		Source   string          `json:"source"`
		Custom   json.RawMessage `json:"custom_details"`
	} `json:"payload"`
}

// EventResult is returned to the Events API caller.
type EventResult struct {
	Status     string `json:"status"`
	DedupKey   string `json:"dedup_key"`
	IncidentID string `json:"incident_id"`
}

// ProcessEvent is the Events API entry point. It routes an inbound event to a
// service by integration key and applies the trigger/ack/resolve action with
// deduplication (one open incident per service + dedup_key).
func (s *IncidentService) ProcessEvent(ctx context.Context, req EventRequest) (*EventResult, error) {
	svc, err := s.services.FindByIntegrationKey(ctx, req.RoutingKey)
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, apperror.NotFound("unknown routing key")
	}

	switch req.EventAction {
	case "trigger":
		return s.triggerFromEvent(ctx, svc, req)
	case "acknowledge":
		return s.transitionFromEvent(ctx, svc, req, "acknowledge")
	case "resolve":
		return s.transitionFromEvent(ctx, svc, req, "resolve")
	default:
		return nil, apperror.BadRequest("invalid event_action")
	}
}

func (s *IncidentService) triggerFromEvent(ctx context.Context, svc *model.Service, req EventRequest) (*EventResult, error) {
	if svc.Status == "disabled" {
		return nil, apperror.BadRequest("service is disabled")
	}
	dedup := req.DedupKey
	if dedup == "" {
		dedup = uuid.NewString()
	}
	// A service under maintenance accepts events but suppresses new incidents.
	if svc.Status == "maintenance" {
		return &EventResult{Status: "suppressed", DedupKey: dedup}, nil
	}

	existing, err := s.incidents.FindOpenByDedup(ctx, svc.ID, dedup)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		_ = s.incidents.AppendLog(ctx, &model.IncidentLogEntry{
			IncidentID: existing.ID, Type: "trigger",
			Message: "Duplicate trigger suppressed (same dedup_key, incident still open)",
		})
		return &EventResult{Status: "deduplicated", DedupKey: dedup, IncidentID: existing.ID}, nil
	}

	title := strings.TrimSpace(req.Payload.Summary)
	if title == "" {
		title = "Incident tanpa judul"
	}
	details, _ := json.Marshal(req.Payload)

	inc := &model.Incident{
		ID:                 uuid.NewString(),
		OrganizationID:     svc.OrganizationID,
		ServiceID:          svc.ID,
		Title:              title,
		Status:             model.IncidentTriggered,
		Urgency:            urgencyFromSeverity(req.Payload.Severity),
		DedupKey:           dedup,
		Details:            details,
		EscalationPolicyID: svc.EscalationPolicyID,
		EscalationLevel:    0,
		TriggeredAt:        time.Now(),
	}
	if err := s.incidents.Create(ctx, inc); err != nil {
		return nil, err
	}
	_ = s.incidents.AppendLog(ctx, &model.IncidentLogEntry{
		IncidentID: inc.ID, Type: "trigger", Message: "Incident created: " + title,
	})
	s.notifier.NotifySlack(ctx, inc, slackURL(svc), "🔴 Triggered: "+title)

	// Page the first escalation level immediately.
	if err := s.escalateCurrentLevel(ctx, inc); err != nil {
		return nil, err
	}
	return &EventResult{Status: "triggered", DedupKey: dedup, IncidentID: inc.ID}, nil
}

func (s *IncidentService) transitionFromEvent(ctx context.Context, svc *model.Service, req EventRequest, action string) (*EventResult, error) {
	if req.DedupKey == "" {
		return nil, apperror.BadRequest("dedup_key is required for " + action)
	}
	inc, err := s.incidents.FindOpenByDedup(ctx, svc.ID, req.DedupKey)
	if err != nil {
		return nil, err
	}
	if inc == nil {
		return nil, apperror.NotFound("no open incident for that dedup_key")
	}
	if action == "acknowledge" {
		if err := s.acknowledge(ctx, inc, nil, "Events API"); err != nil {
			return nil, err
		}
		return &EventResult{Status: "acknowledged", DedupKey: req.DedupKey, IncidentID: inc.ID}, nil
	}
	if err := s.resolve(ctx, inc, nil, "Events API"); err != nil {
		return nil, err
	}
	return &EventResult{Status: "resolved", DedupKey: req.DedupKey, IncidentID: inc.ID}, nil
}

// --- UI-facing operations ---

func (s *IncidentService) List(ctx context.Context, orgID string, f repository.IncidentFilter) ([]model.Incident, error) {
	return s.incidents.ListByOrg(ctx, orgID, f)
}

func (s *IncidentService) Get(ctx context.Context, id, orgID string) (*model.Incident, error) {
	inc, err := s.incidents.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if inc == nil {
		return nil, apperror.NotFound("incident not found")
	}
	if inc.Log, err = s.incidents.ListLog(ctx, inc.ID); err != nil {
		return nil, err
	}
	return inc, nil
}

func (s *IncidentService) Acknowledge(ctx context.Context, id, orgID, userID string) (*model.Incident, error) {
	inc, err := s.Get(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	uid := userID
	if err := s.acknowledge(ctx, inc, &uid, "web"); err != nil {
		return nil, err
	}
	return s.Get(ctx, id, orgID)
}

func (s *IncidentService) Resolve(ctx context.Context, id, orgID, userID string) (*model.Incident, error) {
	inc, err := s.Get(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	uid := userID
	if err := s.resolve(ctx, inc, &uid, "web"); err != nil {
		return nil, err
	}
	return s.Get(ctx, id, orgID)
}

// --- shared transitions ---

func (s *IncidentService) acknowledge(ctx context.Context, inc *model.Incident, userID *string, source string) error {
	if inc.Status == model.IncidentResolved {
		return apperror.BadRequest("incident is already resolved")
	}
	now := time.Now()
	inc.Status = model.IncidentAcknowledged
	inc.AcknowledgedAt = &now
	// Ack halts escalation — unless the service sets an ack timeout, in which
	// case the incident re-escalates if it isn't resolved within that window.
	inc.NextEscalationAt = nil
	if svc, _ := s.services.FindByID(ctx, inc.ServiceID, inc.OrganizationID); svc != nil && svc.AckTimeoutMinutes != nil {
		deadline := now.Add(time.Duration(*svc.AckTimeoutMinutes) * time.Minute)
		inc.NextEscalationAt = &deadline
	}
	if userID != nil {
		inc.AssignedUserID = userID
	}
	if err := s.incidents.Update(ctx, inc); err != nil {
		return err
	}
	return s.incidents.AppendLog(ctx, &model.IncidentLogEntry{
		IncidentID: inc.ID, Type: "acknowledge", Message: "Acknowledged via " + source, UserID: userID,
	})
}

func (s *IncidentService) resolve(ctx context.Context, inc *model.Incident, userID *string, source string) error {
	now := time.Now()
	inc.Status = model.IncidentResolved
	inc.ResolvedAt = &now
	inc.NextEscalationAt = nil
	if err := s.incidents.Update(ctx, inc); err != nil {
		return err
	}
	svc, _ := s.services.FindByID(ctx, inc.ServiceID, inc.OrganizationID)
	if svc != nil {
		s.notifier.NotifySlack(ctx, inc, slackURL(svc), "✅ Resolved: "+inc.Title)
	}
	return s.incidents.AppendLog(ctx, &model.IncidentLogEntry{
		IncidentID: inc.ID, Type: "resolve", Message: "Resolved via " + source, UserID: userID,
	})
}

// --- escalation engine internals (also used by the background worker) ---

// escalateCurrentLevel pages the targets of the incident's current escalation
// rule and arms next_escalation_at for the following level.
func (s *IncidentService) escalateCurrentLevel(ctx context.Context, inc *model.Incident) error {
	if inc.EscalationPolicyID == nil {
		inc.NextEscalationAt = nil
		if err := s.incidents.Update(ctx, inc); err != nil {
			return err
		}
		return s.incidents.AppendLog(ctx, &model.IncidentLogEntry{
			IncidentID: inc.ID, Type: "escalate", Message: "No escalation policy — nobody paged",
		})
	}
	policy, err := s.policies.FindByID(ctx, *inc.EscalationPolicyID, inc.OrganizationID)
	if err != nil {
		return err
	}
	if policy == nil || len(policy.Rules) == 0 || inc.EscalationLevel >= len(policy.Rules) {
		inc.NextEscalationAt = nil
		return s.incidents.Update(ctx, inc)
	}

	rule := policy.Rules[inc.EscalationLevel]
	users := s.resolveTargets(ctx, inc.OrganizationID, rule)
	if len(users) > 0 {
		inc.AssignedUserID = &users[0].ID
	}
	voice := inc.Urgency == "high"
	names := make([]string, 0, len(users))
	for _, u := range users {
		s.notifier.NotifyUser(ctx, u, inc, voice)
		names = append(names, u.Name)
	}

	next := time.Now().Add(time.Duration(rule.DelayMinutes) * time.Minute)
	inc.NextEscalationAt = &next
	if err := s.incidents.Update(ctx, inc); err != nil {
		return err
	}

	msg := fmt.Sprintf("Escalation level %d — paged: %s", inc.EscalationLevel+1, strings.Join(names, ", "))
	if len(names) == 0 {
		msg = fmt.Sprintf("Escalation level %d — no matching on-call", inc.EscalationLevel+1)
	}
	return s.incidents.AppendLog(ctx, &model.IncidentLogEntry{
		IncidentID: inc.ID, Type: "escalate", Message: msg,
	})
}

// Escalate advances one triggered incident to its next level (or repeats /
// exhausts the policy). Called by the background worker when next_escalation_at
// passes.
func (s *IncidentService) Escalate(ctx context.Context, inc *model.Incident) error {
	if inc.Status == model.IncidentResolved || inc.EscalationPolicyID == nil {
		inc.NextEscalationAt = nil
		return s.incidents.Update(ctx, inc)
	}
	policy, err := s.policies.FindByID(ctx, *inc.EscalationPolicyID, inc.OrganizationID)
	if err != nil {
		return err
	}
	if policy == nil || len(policy.Rules) == 0 {
		inc.NextEscalationAt = nil
		return s.incidents.Update(ctx, inc)
	}

	// A due acknowledged incident means the ack timed out: re-open it and keep
	// climbing the escalation chain.
	if inc.Status == model.IncidentAcknowledged {
		inc.Status = model.IncidentTriggered
		_ = s.incidents.AppendLog(ctx, &model.IncidentLogEntry{
			IncidentID: inc.ID, Type: "escalate", Message: "Acknowledgement timed out — re-escalating",
		})
	}

	if inc.EscalationLevel+1 < len(policy.Rules) {
		inc.EscalationLevel++
		return s.escalateCurrentLevel(ctx, inc)
	}
	// Past the last rule: repeat from the top if the policy allows, else stop.
	if inc.EscalationRepeats < policy.RepeatCount {
		inc.EscalationRepeats++
		inc.EscalationLevel = 0
		return s.escalateCurrentLevel(ctx, inc)
	}
	inc.NextEscalationAt = nil
	if err := s.incidents.Update(ctx, inc); err != nil {
		return err
	}
	return s.incidents.AppendLog(ctx, &model.IncidentLogEntry{
		IncidentID: inc.ID, Type: "escalate", Message: "Escalation policy exhausted — incident still unacknowledged",
	})
}

// AutoResolve closes an incident that has passed its service's auto-resolve window.
func (s *IncidentService) AutoResolve(ctx context.Context, inc *model.Incident) error {
	return s.resolve(ctx, inc, nil, "auto-resolve")
}

func (s *IncidentService) DueForEscalation(ctx context.Context, now time.Time) ([]model.Incident, error) {
	return s.incidents.ListDueForEscalation(ctx, now)
}

func (s *IncidentService) DueForAutoResolve(ctx context.Context, now time.Time) ([]model.Incident, error) {
	return s.incidents.ListDueForAutoResolve(ctx, now)
}

// resolveTargets turns an escalation rule's targets into concrete users,
// resolving schedule targets to whoever is on-call now, de-duplicated.
func (s *IncidentService) resolveTargets(ctx context.Context, orgID string, rule model.EscalationRule) []*model.User {
	seen := map[string]bool{}
	var users []*model.User
	add := func(userID string) {
		if userID == "" || seen[userID] {
			return
		}
		u, err := s.users.FindByID(ctx, userID)
		if err != nil || u == nil || u.OrganizationID != orgID {
			return
		}
		seen[userID] = true
		users = append(users, u)
	}
	for _, t := range rule.Targets {
		switch t.TargetType {
		case model.TargetUser:
			add(t.TargetID)
		case model.TargetSchedule:
			sch, err := s.schedules.FindByID(ctx, t.TargetID, orgID)
			if err == nil && sch != nil {
				add(ResolveOnCall(sch, time.Now()))
			}
		}
	}
	return users
}

func urgencyFromSeverity(sev string) string {
	switch strings.ToLower(sev) {
	case "critical", "error", "high":
		return "high"
	case "warning", "info", "low":
		return "low"
	default:
		return "high"
	}
}

func slackURL(svc *model.Service) string {
	if svc.SlackWebhookURL != nil {
		return *svc.SlackWebhookURL
	}
	return ""
}
