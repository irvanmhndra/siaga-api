// Package contract_test boots the real API against a throwaway database and
// checks response shapes that the web clients depend on.
package contract_test

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/irvanmhndra/siaga-api/config"
	"github.com/irvanmhndra/siaga-api/internal/app"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v5"
	_ "github.com/lib/pq"
)

var api *echo.Echo

func TestMain(m *testing.M) {
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		log.Print("contract tests skipped: TEST_DATABASE_URL is not set")
		os.Exit(0)
	}
	os.Exit(run(m, base))
}

func run(m *testing.M, base string) int {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	u, drop, err := createDatabase(base)
	if err != nil {
		log.Printf("create contract database: %v", err)
		return 1
	}
	defer drop()

	// app.New runs migrations from a path relative to the module root.
	if err := os.Chdir(moduleRoot()); err != nil {
		log.Printf("chdir: %v", err)
		return 1
	}
	cfg := config.Load()
	password, _ := u.User.Password()
	cfg.Postgres = config.PostgresConfig{
		Host:     u.Hostname(),
		Port:     u.Port(),
		User:     u.User.Username(),
		Password: password,
		DB:       u.Path[1:],
		SSLMode:  u.Query().Get("sslmode"),
	}
	a, err := app.New(cfg)
	if err != nil {
		log.Printf("start app: %v", err)
		return 1
	}
	defer a.Close()
	api = a.Echo()

	return m.Run()
}

// createDatabase creates a uniquely named database on the server behind base
// and returns its URL plus a func that drops it again.
func createDatabase(base string) (*url.URL, func(), error) {
	admin, err := sqlx.Connect("postgres", base)
	if err != nil {
		return nil, nil, err
	}
	name := fmt.Sprintf("siaga_contract_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		_ = admin.Close()
		return nil, nil, err
	}
	u, err := url.Parse(base)
	if err != nil {
		_ = admin.Close()
		return nil, nil, err
	}
	u.Path = "/" + name
	drop := func() {
		if _, err := admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)"); err != nil {
			log.Printf("drop %s: %v", name, err)
		}
		_ = admin.Close()
	}
	return u, drop, nil
}

func moduleRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}
