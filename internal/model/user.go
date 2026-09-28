package model

import "time"

type User struct {
	ID             string    `db:"id" json:"id"`
	OrganizationID string    `db:"organization_id" json:"organization_id"`
	Name           string    `db:"name" json:"name"`
	Email          string    `db:"email" json:"email"`
	PasswordHash   string    `db:"password_hash" json:"-"`
	Phone          *string   `db:"phone" json:"phone,omitempty"`
	Timezone       string    `db:"timezone" json:"timezone"`
	NotifyEmail    bool      `db:"notify_email" json:"notify_email"`
	NotifySMS      bool      `db:"notify_sms" json:"notify_sms"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type UserSession struct {
	ID               string    `db:"id"`
	UserID           string    `db:"user_id"`
	TokenHash        string    `db:"token_hash"`         // sessiontoken.Hash; never the raw token
	RefreshTokenHash string    `db:"refresh_token_hash"` // sessiontoken.Hash; never the raw token
	ExpiresAt        time.Time `db:"expires_at"`
	RefreshExpiresAt time.Time `db:"refresh_expires_at"`
	CreatedAt        time.Time `db:"created_at"`
}
