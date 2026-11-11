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

// NullableID stores the ID 0 (nobody, nothing) as NULL
func NullableID(id int64) *int64 {
	if id == 0 {
		return nil
	}
	return &id
}

// IDOrZero reads a nullable ID column, NULL as 0
func IDOrZero(id *int64) int64 {
	if id == nil {
		return 0
	}
	return *id
}
