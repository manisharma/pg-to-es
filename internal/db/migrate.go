package db

import (
	"embed"
	"errors"
	"fmt"
	"pg-to-es/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const version = 1

//go:embed migrations/*.sql
var files embed.FS

func Migrate(cfg config.Pg) error {
	driver, err := iofs.New(files, "migrations")
	if err != nil {
		return fmt.Errorf("iofs.New() failed, err: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", driver, cfg.String())
	if err != nil {
		return fmt.Errorf("migrate.NewWithSourceInstance() failed, err: %w", err)
	}
	if err := m.Migrate(version); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("migrate() failed, err: %w", err)
	}
	return nil
}
