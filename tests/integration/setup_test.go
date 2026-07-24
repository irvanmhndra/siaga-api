package integration_test

import (
	"log"
	"os"
	"testing"

	"github.com/irvanmhndra/siaga-api/tests/testutil"
)

var testDB *testutil.TestDB

func TestMain(m *testing.M) {
	db, err := testutil.NewTestDB()
	if err != nil {
		// No database available (e.g. local machine without Postgres): skip the
		// whole integration suite rather than failing `go test ./...`.
		log.Printf("integration tests skipped: %v", err)
		os.Exit(0)
	}
	if err := db.Migrate(); err != nil {
		log.Fatalf("migrate test db: %v", err)
	}
	testDB = db
	code := m.Run()
	_ = db.Close()
	os.Exit(code)
}

// reset gives each test a clean schema.
func reset(t *testing.T) {
	t.Helper()
	if testDB == nil {
		t.Skip("no test database")
	}
	if err := testDB.Truncate(); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}
