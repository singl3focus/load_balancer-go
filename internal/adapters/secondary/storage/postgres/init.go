package postgres

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/singl3focus/load_balancer-go/internal/domain"
	"github.com/singl3focus/load_balancer-go/pkg/logger"
)

type Database struct {
	logger logger.Logger
	db     *sqlx.DB
}

func NewDB(url string, lg logger.Logger) domain.Storage {
	db, err := sqlx.Connect("postgres", url)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return &Database{
		db:     db,
		logger: lg,
	}
}
