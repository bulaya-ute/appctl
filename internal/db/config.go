package db

import "database/sql"

func GetConfig(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query(`SELECT key, value FROM config`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cfg := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		cfg[k] = v
	}
	return cfg, rows.Err()
}

func GetConfigValue(db *sql.DB, key, fallback string) string {
	var v string
	err := db.QueryRow(`SELECT value FROM config WHERE key = ?`, key).Scan(&v)
	if err != nil {
		return fallback
	}
	return v
}

func SetConfigValue(db *sql.DB, key, value string) error {
	_, err := db.Exec(
		`INSERT INTO config (key, value) VALUES (?,?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}
