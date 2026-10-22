package model

import "time"

// Target types for an escalation target.
const (
	TargetUser     = "user"
	TargetSchedule = "schedule"
)

// EscalationPolicy is an ordered set of rules. When an incident is not
// acknowledged within a rule's delay, it advances to the next rule. After the
// last rule it repeats up to RepeatCount times.
type EscalationPolicy struct {
	ID             string    `db:"id" json:"id"`
	OrganizationID string    `db:"organization_id" json:"organization_id"`
	Name           string    `db:"name" json:"name"`
	RepeatCount    int       `db:"repeat_count" json:"repeat_count"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`

	Rules []EscalationRule `db:"-" json:"rules,omitzero"`
}

// EscalationRule fires its targets, then waits DelayMinutes before the next rule.
type EscalationRule struct {
	ID                 string    `db:"id" json:"id"`
	EscalationPolicyID string    `db:"escalation_policy_id" json:"escalation_policy_id"`
	Position           int       `db:"position" json:"position"`
	DelayMinutes       int       `db:"delay_minutes" json:"delay_minutes"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`

	Targets []EscalationTarget `db:"-" json:"targets,omitzero"`
}

// EscalationTarget points a rule at either a user directly or an on-call schedule.
type EscalationTarget struct {
	ID              string    `db:"id" json:"id"`
	EscalationRuleID string   `db:"escalation_rule_id" json:"escalation_rule_id"`
	TargetType      string    `db:"target_type" json:"target_type"`
	TargetID        string    `db:"target_id" json:"target_id"`
	Position        int       `db:"position" json:"position"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}
