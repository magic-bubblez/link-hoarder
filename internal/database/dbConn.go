package database

import (
	"context" //used here for timeouts and cancellations
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connection() (*pgxpool.Pool, error) {
	db_url := "postgres://admin:secret@localhost:5432/link-hoarder"

	config, err := pgxpool.ParseConfig(db_url) //convert db url to config struct which checks if url format is valid
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnIdleTime = 5 * time.Minute

	connpool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := connpool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("database unreachable: %w", err)
	}

	DB = connpool
	return connpool, nil
}
