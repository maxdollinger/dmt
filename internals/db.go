package internals

import (
	"context"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDb(ctx context.Context, databaseURL string) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	return pool
}
