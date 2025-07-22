package main

import (
	"context"
	"dmt/config"
	"dmt/internals/middleware"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5"
)

func main() {
	apiKey := config.GetAPIKey()
	databaseURL := config.GetDatabaseURL()

	// Connect to database
	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	app := fiber.New()

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middleware.KeyAuthMiddleware(apiKey))

	app.Get("/", func(c *fiber.Ctx) error {
		var greeting string
		err := conn.QueryRow(c.Context(), "select 'Hello, World!'").Scan(&greeting)
		if err != nil {
			return c.Status(500).SendString("Database error: " + err.Error())
		}
		return c.SendString(greeting)
	})

	port := config.GetPort()
	log.Fatal(app.Listen(":" + port))
}
