package model

import (
	"time"

	"github.com/lib/pq"
)

// Schedule is a named on-call calendar. Who is on-call at any instant is
// computed from its layers (rotations) with overrides taking precedence.
type Schedule struct {
	ID             string    `db:"id" json:"id"`
	OrganizationID string    `db:"organization_id" json:"organization_id"`
	Name           string    `db:"name" json:"name"`
	Timezone       string    `db:"timezone" json:"timezone"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`

	Layers    []ScheduleLayer    `db:"-" json:"layers,omitempty"`
	Overrides []ScheduleOverride `db:"-" json:"overrides,omitempty"`
}

// ScheduleLayer is one rotation ring. user_ids rotate every
// rotation_length_days, handing off at handoff_time, starting from start_at.
type ScheduleLayer struct {
	ID                 string         `db:"id" json:"id"`
	ScheduleID         string         `db:"schedule_id" json:"schedule_id"`
	Name               string         `db:"name" json:"name"`
	RotationType       string         `db:"rotation_type" json:"rotation_type"`
	RotationLengthDays int            `db:"rotation_length_days" json:"rotation_length_days"`
	HandoffTime        string         `db:"handoff_time" json:"handoff_time"`
	StartAt            time.Time      `db:"start_at" json:"start_at"`
	UserIDs            pq.StringArray `db:"user_ids" json:"user_ids"`
	Position           int            `db:"position" json:"position"`
	CreatedAt          time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at" json:"updated_at"`
}

// ScheduleOverride pins a specific user on-call for a window, winning over the
// underlying rotation.
type ScheduleOverride struct {
	ID         string    `db:"id" json:"id"`
	ScheduleID string    `db:"schedule_id" json:"schedule_id"`
	UserID     string    `db:"user_id" json:"user_id"`
	StartAt    time.Time `db:"start_at" json:"start_at"`
	EndAt      time.Time `db:"end_at" json:"end_at"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
