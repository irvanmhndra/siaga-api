package service

import (
	"context"
	"log/slog"
	"time"
)

// EscalationEngine is the background worker that drives time-based escalation.
// Every tick it advances incidents whose next_escalation_at has passed and
// auto-resolves incidents past their service's auto-resolve window.
type EscalationEngine struct {
	incidents *IncidentService
	interval  time.Duration
}

func NewEscalationEngine(incidents *IncidentService) *EscalationEngine {
	return &EscalationEngine{incidents: incidents, interval: 15 * time.Second}
}

// Start runs the engine loop until ctx is cancelled.
func (e *EscalationEngine) Start(ctx context.Context) {
	slog.Info("Escalation engine started", "interval", e.interval.String())
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("Escalation engine stopped")
			return
		case <-ticker.C:
			e.tick(ctx)
		}
	}
}

func (e *EscalationEngine) tick(ctx context.Context) {
	now := time.Now()

	due, err := e.incidents.DueForEscalation(ctx, now)
	if err != nil {
		slog.Error("engine: list due for escalation", "error", err)
	}
	for i := range due {
		inc := due[i]
		if err := e.incidents.Escalate(ctx, &inc); err != nil {
			slog.Error("engine: escalate", "incident", inc.ID, "error", err)
		}
	}

	autoRes, err := e.incidents.DueForAutoResolve(ctx, now)
	if err != nil {
		slog.Error("engine: list due for auto-resolve", "error", err)
	}
	for i := range autoRes {
		inc := autoRes[i]
		if err := e.incidents.AutoResolve(ctx, &inc); err != nil {
			slog.Error("engine: auto-resolve", "incident", inc.ID, "error", err)
		}
	}
}
