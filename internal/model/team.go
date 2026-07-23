package model

import "time"

type Team struct {
	ID             string    `db:"id" json:"id"`
	OrganizationID string    `db:"organization_id" json:"organization_id"`
	Name           string    `db:"name" json:"name"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`

	Members []User `db:"-" json:"members,omitempty"`
}

type TeamMember struct {
	TeamID string `db:"team_id" json:"team_id"`
	UserID string `db:"user_id" json:"user_id"`
}
