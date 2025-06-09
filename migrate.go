package migrate

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
)

const (
	DefaultTableName = "schema_migrations"
	SuffixUp         = "up.sql"
	SuffixDown       = "down.sql"
)

type migrate struct {
	path      string
	tableName string
	db        *sql.DB
}

func New(db *sql.DB, path string, tableName ...string) *migrate {
	tbName := DefaultTableName
	if len(tableName) > 0 && tableName[0] != "" {
		tbName = tableName[0]
	}

	return &migrate{
		db:        db,
		path:      path,
		tableName: tbName,
	}
}

func (m *migrate) Up() error {
	if err := m.initializeTable(); err != nil {
		return fmt.Errorf("failed to initialize migration table %s: %w", m.tableName, err)
	}
	files, err := findMigrationFiles(m.path, SuffixUp, false)
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}
	current, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}
	for _, file := range files {
		version := strings.Split(file, "_")[0]
		if strings.Compare(version, current) <= 0 {
			continue
		}

		query, err := readMigrationFile(m.path, file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if _, err := m.db.Exec(string(query)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		if err := m.addMigration(version); err != nil {
			return fmt.Errorf("failed to add migration version %s: %w", version, err)
		}
	}

	slog.Info("migration up to date", "version", current)
	return nil
}

func (m *migrate) Down() error {
	files, err := findMigrationFiles(m.path, SuffixDown, true)
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}
	current, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}
	for _, file := range files {
		version := strings.Split(file, "_")[0]
		if strings.Compare(version, current) > 0 {
			continue
		}

		query, err := readMigrationFile(m.path, file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if _, err := m.db.Exec(string(query)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		if err := m.removeMigration(version); err != nil {
			return fmt.Errorf("failed to remove migration version %s: %w", version, err)
		}
	}
	return nil
}

func (m *migrate) Version() (string, error) {
	var version string
	err := m.db.QueryRow(fmt.Sprintf("SELECT version FROM %s ORDER BY version DESC LIMIT 1", m.tableName)).Scan(&version)
	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return version, nil
}

func (m *migrate) initializeTable() error {
	_, err := m.db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (version TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)", m.tableName))
	if err != nil {
		return err
	}
	return nil
}

func (m *migrate) addMigration(version string) error {
	_, err := m.db.Exec(fmt.Sprintf("INSERT INTO %s (version) VALUES (?)", m.tableName), version)
	if err != nil {
		return err
	}
	return nil
}

func (m *migrate) removeMigration(version string) error {
	_, err := m.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE version = ?", m.tableName), version)
	if err != nil {
		return err
	}
	return nil
}
