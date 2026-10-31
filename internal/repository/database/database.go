package database

import (
	"context"

	"github.com/faruq-rahmadani/masjid-digital/internal/config"
	"github.com/faruq-rahmadani/masjid-digital/pkg/database"
)

type DatabaseClient struct {
	database.Querier
}

func NewDatabaseClient(cfg config.Schema) *DatabaseClient {
	db, err := database.Open(context.Background(), cfg.App.Database.ConnConfig())
	if err != nil {
		panic(err)
	}

	return &DatabaseClient{Querier: db}
}
