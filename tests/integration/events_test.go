package integration_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/irvanmhndra/siaga-api/config"
	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/irvanmhndra/siaga-api/internal/repository"
	"github.com/irvanmhndra/siaga-api/internal/repository/postgres"
	"github.com/irvanmhndra/siaga-api/internal/service"
)

// env bundles the service graph wired to the test database.
type env struct {
	auth       *service.AuthService
	escalation *service.EscalationService
	services   *service.ServiceService
	schedules  *service.ScheduleService
	incidents  *service.IncidentService
	notifRepo  interface {
		ListByUser(ctx context.Context, userID string, unreadOnly bool, limit int) ([]model.Notification, error)
	}
}

func newEnv(t *testing.T) *env {
	t.Helper()
	reset(t)
	db := testDB.DB

	orgRepo := postgres.NewOrganizationRepository(db)
	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewUserSessionRepository(db)
	policyRepo := postgres.NewEscalationPolicyRepository(db)
	serviceRepo := postgres.NewServiceRepository(db)
	scheduleRepo := postgres.NewScheduleRepository(db)
	incidentRepo := postgres.NewIncidentRepository(db)
	notifRepo := postgres.NewNotificationRepository(db)

	// Empty SMTP/Twilio/Slack config → external channels disabled, in-app only.
	cfg := &config.Config{App: config.AppConfig{BaseURL: "http://test"}}
	notifier := service.NewNotifier(cfg, notifRepo)

	return &env{
		auth:       service.NewAuthService(orgRepo, userRepo, sessionRepo),
		escalation: service.NewEscalationService(policyRepo),
		services:   service.NewServiceService(serviceRepo, policyRepo),
		schedules:  service.NewScheduleService(scheduleRepo),
		incidents:  service.NewIncidentService(incidentRepo, serviceRepo, policyRepo, scheduleRepo, userRepo, notifier),
		notifRepo:  notifRepo,
	}
}

// register creates an org + owner and returns (userID, orgID).
func (e *env) register(t *testing.T, ctx context.Context, email string) (string, string) {
	t.Helper()
	res, err := e.auth.Register(ctx, service.RegisterRequest{Name: "Ops", Email: email, Password: "secret123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	return res.User.ID, res.User.OrganizationID
}

func userTargetPolicy(t *testing.T, ctx context.Context, e *env, orgID, userID string, delays ...int) *model.EscalationPolicy {
	t.Helper()
	rules := make([]service.RuleRequest, len(delays))
	for i, d := range delays {
		rules[i] = service.RuleRequest{
			DelayMinutes: d,
			Targets:      []service.TargetRequest{{TargetType: "user", TargetID: userID}},
		}
	}
	pol, err := e.escalation.Create(ctx, orgID, service.PolicyRequest{Name: "P", RepeatCount: 1, Rules: rules})
	if err != nil {
		t.Fatalf("create policy: %v", err)
	}
	return pol
}

func triggerEvent(routingKey, dedup, summary, severity string) service.EventRequest {
	req := service.EventRequest{RoutingKey: routingKey, EventAction: "trigger", DedupKey: dedup}
	req.Payload.Summary = summary
	req.Payload.Severity = severity
	return req
}

func TestEventsAPI_Lifecycle(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	uid, orgID := e.register(t, ctx, "ops@example.com")
	pol := userTargetPolicy(t, ctx, e, orgID, uid, 30)

	svc, err := e.services.Create(ctx, orgID, service.ServiceRequest{Name: "Checkout", EscalationPolicyID: &pol.ID})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	key := svc.IntegrationKey

	// Trigger.
	res, err := e.incidents.ProcessEvent(ctx, triggerEvent(key, "disk-1", "Disk full", "critical"))
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if res.Status != "triggered" {
		t.Fatalf("want triggered, got %q", res.Status)
	}
	incID := res.IncidentID

	inc, err := e.incidents.Get(ctx, incID, orgID)
	if err != nil {
		t.Fatalf("get incident: %v", err)
	}
	if inc.Status != model.IncidentTriggered {
		t.Errorf("status: want triggered, got %q", inc.Status)
	}
	if inc.Urgency != "high" {
		t.Errorf("urgency: want high (from critical), got %q", inc.Urgency)
	}
	if inc.AssignedUserID == nil || *inc.AssignedUserID != uid {
		t.Errorf("assigned: want %s, got %v", uid, inc.AssignedUserID)
	}
	if len(inc.Log) < 2 {
		t.Errorf("timeline: want >=2 entries (trigger+escalate), got %d", len(inc.Log))
	}

	// In-app notification recorded for the on-call user.
	notifs, err := e.notifRepo.ListByUser(ctx, uid, false, 10)
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if len(notifs) == 0 {
		t.Error("expected an in-app notification for the paged user")
	}

	// Duplicate trigger dedups to the same incident.
	dup, err := e.incidents.ProcessEvent(ctx, triggerEvent(key, "disk-1", "Disk full again", "critical"))
	if err != nil {
		t.Fatalf("duplicate trigger: %v", err)
	}
	if dup.Status != "deduplicated" || dup.IncidentID != incID {
		t.Fatalf("dedup: want deduplicated on %s, got %q on %s", incID, dup.Status, dup.IncidentID)
	}

	// Acknowledge via the Events API halts escalation.
	ack := service.EventRequest{RoutingKey: key, EventAction: "acknowledge", DedupKey: "disk-1"}
	if _, err := e.incidents.ProcessEvent(ctx, ack); err != nil {
		t.Fatalf("acknowledge: %v", err)
	}
	inc, _ = e.incidents.Get(ctx, incID, orgID)
	if inc.Status != model.IncidentAcknowledged {
		t.Errorf("after ack: want acknowledged, got %q", inc.Status)
	}
	if inc.NextEscalationAt != nil {
		t.Error("after ack: next_escalation_at should be cleared")
	}

	// Resolve, then a new trigger on the same dedup_key opens a NEW incident.
	resolve := service.EventRequest{RoutingKey: key, EventAction: "resolve", DedupKey: "disk-1"}
	if _, err := e.incidents.ProcessEvent(ctx, resolve); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	reopen, err := e.incidents.ProcessEvent(ctx, triggerEvent(key, "disk-1", "Disk full once more", "critical"))
	if err != nil {
		t.Fatalf("re-trigger: %v", err)
	}
	if reopen.Status != "triggered" {
		t.Errorf("re-trigger after resolve: want triggered, got %q", reopen.Status)
	}
	if reopen.IncidentID == incID {
		t.Error("re-trigger after resolve should open a new incident, not reuse the resolved one")
	}
}

func TestEscalationEngine_Advances(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	uid, orgID := e.register(t, ctx, "ops@example.com")
	// Two levels, delay 0, repeat_count 1.
	pol := userTargetPolicy(t, ctx, e, orgID, uid, 0, 0)

	svc, err := e.services.Create(ctx, orgID, service.ServiceRequest{Name: "Payments", EscalationPolicyID: &pol.ID})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	res, err := e.incidents.ProcessEvent(ctx, triggerEvent(svc.IntegrationKey, "pay-1", "Down", "critical"))
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	incID := res.IncidentID

	advance := func() *model.Incident {
		inc, err := e.incidents.Get(ctx, incID, orgID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if err := e.incidents.Escalate(ctx, inc); err != nil {
			t.Fatalf("escalate: %v", err)
		}
		out, _ := e.incidents.Get(ctx, incID, orgID)
		return out
	}

	// Level starts at 0 after trigger applies the first rule.
	if inc, _ := e.incidents.Get(ctx, incID, orgID); inc.EscalationLevel != 0 {
		t.Fatalf("initial level: want 0, got %d", inc.EscalationLevel)
	}
	if inc := advance(); inc.EscalationLevel != 1 {
		t.Fatalf("after 1st advance: want level 1, got %d", inc.EscalationLevel)
	}
	// Past the last rule: repeats once and wraps back to level 0.
	if inc := advance(); inc.EscalationLevel != 0 || inc.EscalationRepeats != 1 {
		t.Fatalf("after repeat: want level 0 repeats 1, got level %d repeats %d", inc.EscalationLevel, inc.EscalationRepeats)
	}
	_ = advance() // level 0 -> 1
	// Repeat budget exhausted: escalation stops.
	final := advance()
	if final.NextEscalationAt != nil {
		t.Errorf("exhausted policy: next_escalation_at should be nil, got %v", final.NextEscalationAt)
	}
	if final.Status != model.IncidentTriggered {
		t.Errorf("exhausted policy: incident should still be triggered, got %q", final.Status)
	}
}

func TestMaintenance_SuppressesTriggers(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	uid, orgID := e.register(t, ctx, "ops@example.com")
	pol := userTargetPolicy(t, ctx, e, orgID, uid, 30)

	svc, err := e.services.Create(ctx, orgID, service.ServiceRequest{
		Name: "Under Maintenance", Status: "maintenance", EscalationPolicyID: &pol.ID,
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	res, err := e.incidents.ProcessEvent(ctx, triggerEvent(svc.IntegrationKey, "m-1", "Noise", "critical"))
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if res.Status != "suppressed" {
		t.Errorf("maintenance service: want suppressed, got %q", res.Status)
	}
	incidents, err := e.incidents.List(ctx, orgID, repository.IncidentFilter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(incidents) != 0 {
		t.Errorf("maintenance service should open no incidents, got %d", len(incidents))
	}
}

func TestAckTimeout_ReEscalates(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	uid, orgID := e.register(t, ctx, "ops@example.com")
	pol := userTargetPolicy(t, ctx, e, orgID, uid, 0, 0) // two levels

	ackTimeout := 5
	svc, err := e.services.Create(ctx, orgID, service.ServiceRequest{
		Name: "Billing", EscalationPolicyID: &pol.ID, AckTimeoutMinutes: &ackTimeout,
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	res, err := e.incidents.ProcessEvent(ctx, triggerEvent(svc.IntegrationKey, "bill-1", "Billing down", "critical"))
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	incID := res.IncidentID

	// Acknowledge arms the ack-timeout deadline instead of clearing escalation.
	acked, err := e.incidents.Acknowledge(ctx, incID, orgID, uid)
	if err != nil {
		t.Fatalf("acknowledge: %v", err)
	}
	if acked.Status != model.IncidentAcknowledged {
		t.Fatalf("want acknowledged, got %q", acked.Status)
	}
	if acked.NextEscalationAt == nil {
		t.Fatal("ack with a service ack_timeout should arm next_escalation_at")
	}

	// The engine picks it up once the deadline passes — simulate by escalating.
	inc, _ := e.incidents.Get(ctx, incID, orgID)
	if err := e.incidents.Escalate(ctx, inc); err != nil {
		t.Fatalf("escalate: %v", err)
	}
	out, _ := e.incidents.Get(ctx, incID, orgID)
	if out.Status != model.IncidentTriggered {
		t.Errorf("after ack timeout: want re-triggered, got %q", out.Status)
	}
	if out.EscalationLevel != 1 {
		t.Errorf("after ack timeout: want escalation to advance to level 1, got %d", out.EscalationLevel)
	}
	timedOut := false
	for _, entry := range out.Log {
		if strings.Contains(entry.Message, "timed out") {
			timedOut = true
		}
	}
	if !timedOut {
		t.Error("expected a 'timed out' entry in the incident timeline")
	}
}

func TestEscalation_ScheduleTargetResolvesOnCall(t *testing.T) {
	ctx := context.Background()
	e := newEnv(t)
	uid, orgID := e.register(t, ctx, "ops@example.com")

	sched, err := e.schedules.Create(ctx, orgID, service.ScheduleRequest{
		Name: "Primary",
		Layers: []service.LayerRequest{{
			Name:               "L1",
			RotationLengthDays: 7,
			StartAt:            time.Now().Add(-24 * time.Hour),
			UserIDs:            []string{uid},
		}},
	})
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}

	pol, err := e.escalation.Create(ctx, orgID, service.PolicyRequest{
		Name: "Sched", RepeatCount: 0,
		Rules: []service.RuleRequest{{
			DelayMinutes: 30,
			Targets:      []service.TargetRequest{{TargetType: "schedule", TargetID: sched.ID}},
		}},
	})
	if err != nil {
		t.Fatalf("create policy: %v", err)
	}

	svc, err := e.services.Create(ctx, orgID, service.ServiceRequest{Name: "Gateway", EscalationPolicyID: &pol.ID})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	res, err := e.incidents.ProcessEvent(ctx, triggerEvent(svc.IntegrationKey, "gw-1", "Gateway down", "critical"))
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	inc, _ := e.incidents.Get(ctx, res.IncidentID, orgID)
	if inc.AssignedUserID == nil || *inc.AssignedUserID != uid {
		t.Fatalf("schedule target should resolve to on-call user %s, got %v", uid, inc.AssignedUserID)
	}
}
