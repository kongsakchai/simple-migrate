package migrate

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupDatabase() *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}

	return db
}

func TestMigrationUp(t *testing.T) {
	db := setupDatabase()
	defer db.Close()

	m := New(db, "example")

	// Run the Up migration
	if err := m.Up(); err != nil {
		t.Fatalf("Up migration failed: %v", err)
	}

	// Verify the migration was applied
	version, err := m.Version()
	if err != nil {
		t.Fatalf("Failed to get migration version: %v", err)
	}

	if version != "002" {
		t.Errorf("Expected version 0002, got %s", version)
	}
}

func TestMigrationDown(t *testing.T) {
	db := setupDatabase()
	defer db.Close()

	m := New(db, "example")

	// Run the Up migration first
	if err := m.Up(); err != nil {
		t.Fatalf("Up migration failed: %v", err)
	}

	// Now run the Down migration
	if err := m.Down(); err != nil {
		t.Fatalf("Down migration failed: %v", err)
	}

	// Verify the migration was rolled back
	version, err := m.Version()
	if err != nil {
		t.Fatalf("Failed to get migration version: %v", err)
	}

	if version != "" {
		t.Errorf("Expected no migration, got %s", version)
	}
}
