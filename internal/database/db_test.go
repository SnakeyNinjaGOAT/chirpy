package database

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

// Note: This requires a test database. Set TEST_DB_URL environment variable.
func TestQueries(t *testing.T) {
	dbURL := "postgres://user:password@localhost/chirpy_test?sslmode=disable" // adjust as needed
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skip("Test database not available")
	}
	defer db.Close()

	queries := New(db)

	// Test GetChirps
	chirps, err := queries.GetChirps(nil) // assuming context
	if err != nil {
		t.Fatalf("GetChirps failed: %v", err)
	}
	// Just check no error, since DB may be empty
	_ = chirps
}
