package internals

import (
	"context"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
)

func ConnectDb(ctx context.Context, databaseURL string) *pgx.Conn {
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	return conn
}
