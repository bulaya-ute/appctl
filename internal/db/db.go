package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS apps (
	id             TEXT PRIMARY KEY,
	name           TEXT UNIQUE NOT NULL,
	description    TEXT NOT NULL DEFAULT '',
	type           TEXT NOT NULL,
	source         TEXT NOT NULL DEFAULT 'git',
	local_path     TEXT NOT NULL,
	git_repo_url   TEXT NOT NULL DEFAULT '',
	git_token_path TEXT NOT NULL DEFAULT '',
	branch         TEXT NOT NULL DEFAULT 'main',
	service_name   TEXT NOT NULL DEFAULT '',
	binding_port   INTEGER NOT NULL DEFAULT 0,
	domain         TEXT NOT NULL DEFAULT '',
	build_command  TEXT NOT NULL DEFAULT '',
	run_command    TEXT NOT NULL DEFAULT '',
	publish_dir    TEXT NOT NULL DEFAULT '',
	webhook_secret TEXT NOT NULL DEFAULT '',
	created_at     DATETIME NOT NULL,
	updated_at     DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS deployments (
	id           TEXT PRIMARY KEY,
	app_id       TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
	version      TEXT NOT NULL,
	triggered_by TEXT NOT NULL DEFAULT 'manual',
	status       TEXT NOT NULL DEFAULT 'pending',
	log          TEXT NOT NULL DEFAULT '',
	started_at   DATETIME NOT NULL,
	finished_at  DATETIME
);

CREATE TABLE IF NOT EXISTS config (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL
);

INSERT OR IGNORE INTO config (key, value) VALUES
	('caddy_admin_url', 'http://localhost:2019'),
	('appctl_port', '7070');
`

// DefaultPath returns the default SQLite database path.
// Uses /var/lib/appctl when running as root, otherwise ~/.local/share/appctl.
func DefaultPath() string {
	if os.Getuid() == 0 {
		return "/var/lib/appctl/appctl.db"
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "appctl", "appctl.db")
}

// Open opens (or creates) the SQLite database at path, applies the schema, and returns the connection.
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
