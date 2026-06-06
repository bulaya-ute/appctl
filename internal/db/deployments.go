package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

func CreateDeployment(db *sql.DB, d *Deployment) error {
	d.ID = uuid.New().String()
	d.StartedAt = time.Now().UTC()
	_, err := db.Exec(
		`INSERT INTO deployments (id, app_id, version, triggered_by, status, log, started_at)
		VALUES (?,?,?,?,?,?,?)`,
		d.ID, d.AppID, d.Version, string(d.TriggeredBy), string(d.Status),
		d.Log, d.StartedAt.Format(time.RFC3339),
	)
	return err
}

func UpdateDeployment(db *sql.DB, d *Deployment) error {
	var finishedAt any
	if d.FinishedAt != nil {
		finishedAt = d.FinishedAt.Format(time.RFC3339)
	}
	_, err := db.Exec(
		`UPDATE deployments SET status=?, log=?, finished_at=? WHERE id=?`,
		string(d.Status), d.Log, finishedAt, d.ID,
	)
	return err
}

func ListDeployments(db *sql.DB, appID string, limit int) ([]Deployment, error) {
	rows, err := db.Query(
		`SELECT d.id, d.app_id, a.name, d.version, d.triggered_by, d.status, d.log,
		d.started_at, d.finished_at
		FROM deployments d JOIN apps a ON a.id = d.app_id
		WHERE d.app_id = ? ORDER BY d.started_at DESC LIMIT ?`,
		appID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []Deployment
	for rows.Next() {
		var d Deployment
		var startedAt string
		var finishedAt sql.NullString
		err := rows.Scan(
			&d.ID, &d.AppID, &d.AppName, &d.Version, &d.TriggeredBy,
			&d.Status, &d.Log, &startedAt, &finishedAt,
		)
		if err != nil {
			return nil, err
		}
		d.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
		if finishedAt.Valid {
			t, _ := time.Parse(time.RFC3339, finishedAt.String)
			d.FinishedAt = &t
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}
