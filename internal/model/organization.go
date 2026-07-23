package model

import "time"

// Organization is the top-level tenant. Every domain row carries its
// organization_id, and a user belongs to exactly one organization.
type Organization struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
