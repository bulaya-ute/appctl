package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const appCols = `id, name, description, type, source, local_path, git_repo_url,
	git_token_path, branch, service_name, binding_port, domain,
	build_command, run_command, publish_dir, webhook_secret, created_at, updated_at`

func scanApp(row interface{ Scan(...any) error }) (App, error) {
	var a App
	var createdAt, updatedAt string
	err := row.Scan(
		&a.ID, &a.Name, &a.Description, &a.Type, &a.Source, &a.LocalPath,
		&a.GitRepoURL, &a.GitTokenPath, &a.Branch, &a.ServiceName,
		&a.BindingPort, &a.Domain, &a.BuildCommand, &a.RunCommand,
		&a.PublishDir, &a.WebhookSecret, &createdAt, &updatedAt,
	)
	if err != nil {
		return App{}, err
	}
	a.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	a.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return a, nil
}

func ListApps(db *sql.DB) ([]App, error) {
	rows, err := db.Query(`SELECT ` + appCols + ` FROM apps ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []App
	for rows.Next() {
		a, err := scanApp(rows)
		if err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, rows.Err()
}

func GetApp(db *sql.DB, name string) (*App, error) {
	row := db.QueryRow(`SELECT `+appCols+` FROM apps WHERE name = ?`, name)
	a, err := scanApp(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func GetAppByID(db *sql.DB, id string) (*App, error) {
	row := db.QueryRow(`SELECT `+appCols+` FROM apps WHERE id = ?`, id)
	a, err := scanApp(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func CreateApp(db *sql.DB, a *App) error {
	a.ID = uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(
		`INSERT INTO apps (id, name, description, type, source, local_path, git_repo_url,
		git_token_path, branch, service_name, binding_port, domain,
		build_command, run_command, publish_dir, webhook_secret, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		a.ID, a.Name, a.Description, string(a.Type), string(a.Source), a.LocalPath,
		a.GitRepoURL, a.GitTokenPath, a.Branch, a.ServiceName, a.BindingPort, a.Domain,
		a.BuildCommand, a.RunCommand, a.PublishDir, a.WebhookSecret, now, now,
	)
	return err
}

func UpdateApp(db *sql.DB, a *App) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(
		`UPDATE apps SET description=?, type=?, source=?, local_path=?, git_repo_url=?,
		git_token_path=?, branch=?, service_name=?, binding_port=?, domain=?,
		build_command=?, run_command=?, publish_dir=?, webhook_secret=?, updated_at=?
		WHERE id=?`,
		a.Description, string(a.Type), string(a.Source), a.LocalPath, a.GitRepoURL,
		a.GitTokenPath, a.Branch, a.ServiceName, a.BindingPort, a.Domain,
		a.BuildCommand, a.RunCommand, a.PublishDir, a.WebhookSecret, now, a.ID,
	)
	return err
}

func DeleteApp(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM apps WHERE id = ?`, id)
	return err
}
