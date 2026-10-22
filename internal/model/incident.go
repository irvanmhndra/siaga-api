package model

import (
	"encoding/json"
	"time"
)

// Incident statuses.
const (
	IncidentTriggered    = "triggered"
	IncidentAcknowledged = "acknowledged"
	IncidentResolved     = "resolved"
)

// Incident is opened by a trigger event on a service and moves through
// triggered → acknowledged → resolved. At most one open incident exists per
// (service, dedup_key).
type Incident struct {
	ID                 string          `db:"id" json:"id"`
	OrganizationID     string          `db:"organization_id" json:"organization_id"`
	ServiceID          string          `db:"service_id" json:"service_id"`
	Title              string          `db:"title" json:"title"`
	Status             string          `db:"status" json:"status"`
	Urgency            string          `db:"urgency" json:"urgency"`
	DedupKey           string          `db:"dedup_key" json:"dedup_key"`
	Details            json.RawMessage `db:"details" json:"details,omitempty"`
	AssignedUserID     *string         `db:"assigned_user_id" json:"assigned_user_id,omitempty"`
	EscalationPolicyID *string         `db:"escalation_policy_id" json:"escalation_policy_id,omitempty"`
	EscalationLevel    int             `db:"escalation_level" json:"escalation_level"`
	EscalationRepeats  int             `db:"escalation_repeats" json:"escalation_repeats"`
	NextEscalationAt   *time.Time      `db:"next_escalation_at" json:"next_escalation_at,omitempty"`
	TriggeredAt        time.Time       `db:"triggered_at" json:"triggered_at"`
	AcknowledgedAt     *time.Time      `db:"acknowledged_at" json:"acknowledged_at,omitempty"`
	ResolvedAt         *time.Time      `db:"resolved_at" json:"resolved_at,omitempty"`
	CreatedAt          time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time       `db:"updated_at" json:"updated_at"`

	Log []IncidentLogEntry `db:"-" json:"log,omitzero"`
}

// IncidentLogEntry is one timeline event (trigger, ack, resolve, escalate,
// notify, note).
type IncidentLogEntry struct {
	ID         string    `db:"id" json:"id"`
	IncidentID string    `db:"incident_id" json:"incident_id"`
	Type       string    `db:"type" json:"type"`
	Message    string    `db:"message" json:"message"`
	UserID     *string   `db:"user_id" json:"user_id,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
