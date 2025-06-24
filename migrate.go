package migrate

import (
	"bytes"
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

func (m *migrate) SetVersion(version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	if err := m.initializeTable(); err != nil {
		return fmt.Errorf("failed to initialize migration table %s: %w", m.tableName, err)
	}
	current, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	if current < version {
		return m.upToVersion(version)
	} else if current > version {
		return m.downToVersion(version)
	}

	slog.Info("migration already at specified version", "version", version)
	return nil
}

func (m *migrate) Up() error {
	return m.upToVersion()
}

func (m *migrate) upToVersion(v ...string) error {
	files, err := findMigrationFiles(m.path, SuffixUp, false)
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}

	if err := m.initializeTable(); err != nil {
		return fmt.Errorf("failed to initialize migration table %s: %w", m.tableName, err)
	}
	current, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	for _, file := range files {
		fVersion := strings.Split(file, "_")[0]
		if fVersion <= current {
			continue
		}

		query, err := readMigrationFile(m.path, file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if err := m.execMigration(query); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		if err := m.addMigration(fVersion); err != nil {
			return fmt.Errorf("failed to add migration version %s: %w", fVersion, err)
		}

		if len(v) > 0 && fVersion == v[0] {
			slog.Info("migration up to specified version", "version", fVersion)
			return nil
		}

		current = fVersion
	}

	slog.Info("migration up to latest version", "version", current)
	return nil
}

func (m *migrate) Down() error {
	return m.downToVersion()
}

func (m *migrate) downToVersion(v ...string) error {
	files, err := findMigrationFiles(m.path, SuffixDown, true)
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}

	if err := m.initializeTable(); err != nil {
		return fmt.Errorf("failed to initialize migration table %s: %w", m.tableName, err)
	}
	current, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	for _, file := range files {
		fVersion := strings.Split(file, "_")[0]
		if fVersion > current {
			continue
		}

		if len(v) > 0 && fVersion == v[0] {
			slog.Info("migration down to specified version", "version", fVersion)
			return nil
		}

		query, err := readMigrationFile(m.path, file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if err := m.execMigration(query); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		if err := m.removeMigration(fVersion); err != nil {
			return fmt.Errorf("failed to remove migration version %s: %w", fVersion, err)
		}
	}

	slog.Info("migration donw successfully")
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

func (m *migrate) execMigration(query []byte) (errResp error) {
	tx, err := m.db.Begin()
	if err != nil {
		return errResp
	}

	defer func() {
		if errResp == nil {
			return
		}
		if err := tx.Rollback(); err != nil {
			errResp = fmt.Errorf("rollback failed: %v, original error: %w", err, errResp)
			return
		}
		errResp = fmt.Errorf("migration execution failed: %w", errResp)
	}()

	for {
		index := bytes.Index(query, []byte(";"))
		if index == -1 {
			break
		}

		stmt := query[:index+1]
		query = query[index+1:]
		if len(strings.TrimSpace(string(stmt))) == 0 {
			continue
		}

		if _, err := tx.Exec(string(stmt)); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return errResp
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
