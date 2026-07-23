package model

import "time"

// Service is a monitored system. Events sent to its integration_key open and
// resolve incidents, which are routed through its escalation policy.
type Service struct {
	ID                 string    `db:"id" json:"id"`
	OrganizationID     string    `db:"organization_id" json:"organization_id"`
	Name               string    `db:"name" json:"name"`
	Description        *string   `db:"description" json:"description,omitempty"`
	EscalationPolicyID *string   `db:"escalation_policy_id" json:"escalation_policy_id,omitempty"`
	IntegrationKey     string    `db:"integration_key" json:"integration_key"`
	Status             string    `db:"status" json:"status"`
	AutoResolveMinutes *int      `db:"auto_resolve_minutes" json:"auto_resolve_minutes,omitempty"`
	AckTimeoutMinutes  *int      `db:"ack_timeout_minutes" json:"ack_timeout_minutes,omitempty"`
	SlackWebhookURL    *string   `db:"slack_webhook_url" json:"slack_webhook_url,omitempty"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}
