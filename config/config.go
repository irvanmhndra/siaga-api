package config

import "os"

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	App      AppConfig
	Web      WebConfig
	SMTP     SMTPConfig
	Twilio   TwilioConfig
	Slack    SlackConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

// WebConfig points at the built siaga-web SPA. When DistPath is set the API
// serves the SPA from the same origin (single-origin deployment) with an
// index.html fallback for client-side routes.
type WebConfig struct {
	DistPath string
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
	SSLMode  string
}

func (p PostgresConfig) DSN() string {
	return "postgres://" + p.User + ":" + p.Password + "@" + p.Host + ":" + p.Port + "/" + p.DB + "?sslmode=" + p.SSLMode
}

// AppConfig holds application-level settings such as the public web base URL,
// used when composing links inside notifications (email/SMS/Slack).
type AppConfig struct {
	BaseURL string
}

// SMTPConfig configures outbound email notifications. Disabled when unset.
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func (c SMTPConfig) Enabled() bool {
	return c.Host != "" && c.From != ""
}

// TwilioConfig configures SMS and voice-call notifications. Disabled when unset.
type TwilioConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
}

func (c TwilioConfig) Enabled() bool {
	return c.AccountSID != "" && c.AuthToken != "" && c.FromNumber != ""
}

// SlackConfig holds an org-wide fallback webhook. Per-service webhooks (stored
// on the service row) take precedence over this default.
type SlackConfig struct {
	DefaultWebhookURL string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8089"),
			Env:  getEnv("SERVER_ENV", "development"),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "siaga_user"),
			Password: getEnv("POSTGRES_PASSWORD", "siaga_pass"),
			DB:       getEnv("POSTGRES_DB", "siaga_db"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		App: AppConfig{
			BaseURL: getEnv("APP_BASE_URL", "http://localhost:5182"),
		},
		Web: WebConfig{
			DistPath: getEnv("WEB_DIST", ""),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnv("SMTP_PORT", "587"),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
		},
		Twilio: TwilioConfig{
			AccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
			AuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
			FromNumber: getEnv("TWILIO_FROM_NUMBER", ""),
		},
		Slack: SlackConfig{
			DefaultWebhookURL: getEnv("SLACK_DEFAULT_WEBHOOK_URL", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
