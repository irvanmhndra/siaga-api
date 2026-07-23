package model

import "time"

// Notification channels.
const (
	ChannelInApp = "inapp"
	ChannelEmail = "email"
	ChannelSMS   = "sms"
	ChannelVoice = "voice"
	ChannelSlack = "slack"
)

// Notification is a delivery record. In-app notifications are read via the API;
// email/sms/voice/slack are dispatched to external providers.
type Notification struct {
	ID             string     `db:"id" json:"id"`
	OrganizationID string     `db:"organization_id" json:"organization_id"`
	IncidentID     *string    `db:"incident_id" json:"incident_id,omitempty"`
	UserID         *string    `db:"user_id" json:"user_id,omitempty"`
	Channel        string     `db:"channel" json:"channel"`
	Status         string     `db:"status" json:"status"`
	Message        string     `db:"message" json:"message"`
	ReadAt         *time.Time `db:"read_at" json:"read_at,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
}
