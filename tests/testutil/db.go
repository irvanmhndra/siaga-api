// Package testutil provides shared helpers for Siaga's integration tests.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DefaultDSN points at a local Postgres test database. Override it with the
// TEST_DATABASE_DSN environment variable (e.g. a CI Postgres service).
const DefaultDSN = "postgres://siaga_user:siaga_pass@localhost:5432/siaga_test?sslmode=disable"

// appTables lists every table a test may touch, ordered so a single
// TRUNCATE ... CASCADE resets the whole schema between tests.
var appTables = []string{
	"notifications", "incident_log", "incidents", "services",
	"escalation_targets", "escalation_rules", "escalation_policies",
	"schedule_overrides", "schedule_layers", "schedules",
	"team_members", "teams", "user_sessions", "users", "organizations",
}

// TestDB wraps a connection to the integration test database.
type TestDB struct {
	DB  *sqlx.DB
	DSN string
}

// NewTestDB connects to the test database (TEST_DATABASE_DSN or DefaultDSN).
func NewTestDB() (*TestDB, error) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = DefaultDSN
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect test db: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &TestDB{DB: db, DSN: dsn}, nil
}

// Migrate resets the schema and applies all up migrations for a clean slate.
func (t *TestDB) Migrate() error {
	if _, err := t.DB.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		return fmt.Errorf("reset schema: %w", err)
	}
	dir, err := findMigrationsDir()
	if err != nil {
		return err
	}
	m, err := migrate.New("file://"+dir, t.DSN)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

// Truncate wipes all application tables (and resets identities) between tests.
func (t *TestDB) Truncate() error {
	stmt := "TRUNCATE TABLE " + join(appTables, ", ") + " RESTART IDENTITY CASCADE"
	if _, err := t.DB.Exec(stmt); err != nil {
		return fmt.Errorf("truncate: %w", err)
	}
	return nil
}

func (t *TestDB) Close() error { return t.DB.Close() }

func join(ss []string, sep string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += sep
		}
		out += s
	}
	return out
}

// findMigrationsDir walks up from the working directory to the module root
// (where go.mod lives) and returns the sibling migrations/ directory.
func findMigrationsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "migrations"), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from working dir")
		}
		dir = parent
	}
}
