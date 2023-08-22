package db

import (
	"pg-to-es/internal/config"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

func Connect(cfg config.Pg) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", cfg.String())
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(cfg.MaxIdleTimeForConns)
	db.SetConnMaxLifetime(cfg.MaxLifetimeForConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	return db, nil
}
