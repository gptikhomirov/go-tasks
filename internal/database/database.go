package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const dbDriverName = "postgres"

func Connect(databaseURL string) (*sqlx.DB, error) {
	db, err := sqlx.Connect(dbDriverName, databaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)

	return db, nil
}
