package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose"
)

const (
	rateLimitConfigTable = "rate_limit_configs"
)

const (
	migrationsTable = "goose_db_version_balacner"
)

const (
	OptUp      = "up"
	OptUpAll   = "up-all"
	OptDown    = "down"
	OptDownAll = "down-all"
)

func InvokeMigrator(postgresURL, migrationDir, option string) error {
	db, err := sqlx.Open("postgres", postgresURL)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	goose.SetTableName(migrationsTable)

	switch option {
	case OptUp:
		err = goose.UpByOne(db.DB, migrationDir)
	case OptUpAll:
		err = goose.Up(db.DB, migrationDir)
	case OptDown:
		err = goose.Down(db.DB, migrationDir)
	case OptDownAll:
		err = goose.DownTo(db.DB, migrationDir, 0)
	default:
		return fmt.Errorf("migration error: unknown option %s\n", option)
	}
	if err != nil {
		return fmt.Errorf("migration error: %s\n", err.Error())
	}

	return nil
}
