// Package harnessstore provides persistent storage for harness instance registry
// and credential bindings. This is static configuration data - runtime state
// (active sessions, credential slots) lives in llm-bridge-server.
package harnessstore

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store manages harness instance registry and credential bindings.
type Store struct {
	db *sql.DB
}

// Open opens or creates a harness-store database.
func Open(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite pragmas: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		-- Harness instances: deployments of harness types on specific machines
		CREATE TABLE IF NOT EXISTS instances (
			id                      TEXT PRIMARY KEY,
			harness_type            TEXT NOT NULL,
			name                    TEXT NOT NULL,
			host                    TEXT NOT NULL DEFAULT 'localhost',
			transport               TEXT NOT NULL DEFAULT 'local',
			ssh_user                TEXT NOT NULL DEFAULT '',
			ssh_key_path            TEXT NOT NULL DEFAULT '',
			ssh_port                INTEGER NOT NULL DEFAULT 22,
			working_dir             TEXT NOT NULL DEFAULT '',
			max_concurrent_sessions INTEGER NOT NULL DEFAULT 1,
			enabled                 INTEGER NOT NULL DEFAULT 1,
			created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_instances_harness ON instances(harness_type);
		CREATE INDEX IF NOT EXISTS idx_instances_enabled ON instances(enabled);
		CREATE INDEX IF NOT EXISTS idx_instances_host ON instances(host);

		-- Harness type metadata: display info for each harness type
		CREATE TABLE IF NOT EXISTS harness_types (
			name        TEXT PRIMARY KEY,
			label       TEXT NOT NULL DEFAULT '',
			emoji       TEXT NOT NULL DEFAULT '',
			image       TEXT NOT NULL DEFAULT '',
			updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- Credential bindings: which credentials an instance can use, with priority
		CREATE TABLE IF NOT EXISTS instance_credentials (
			instance_id    TEXT NOT NULL,
			credential_id  TEXT NOT NULL,
			priority       INTEGER NOT NULL DEFAULT 0,
			enabled        INTEGER NOT NULL DEFAULT 1,
			PRIMARY KEY (instance_id, credential_id),
			FOREIGN KEY (instance_id) REFERENCES instances(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_instance_creds_instance ON instance_credentials(instance_id);
		CREATE INDEX IF NOT EXISTS idx_instance_creds_credential ON instance_credentials(credential_id);
	`)
	return err
}

// DB returns the underlying database connection for advanced queries.
func (s *Store) DB() *sql.DB {
	return s.db
}
