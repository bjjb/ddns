package main

import (
	"database/sql"
	"net/url"
)

const schema = `
CREATE TABLE IF NOT EXISTS settings (
	key   TEXT PRIMARY KEY,
	value TEXT
);`

func openDB(s string) (*sql.DB, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	scheme := u.Scheme
	switch scheme {
	case "file":
		scheme = "sqlite3"
	}
	db, err := sql.Open(scheme, s)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	return db, nil
}
