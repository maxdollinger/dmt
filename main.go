package main

import (
	"context"
	"dmt/config"
	"dmt/internals/middleware"
	"dmt/pkg/device"
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

	deviceService := device.NewDeviceService(conn)

	app.Post("/device", deviceService.CreateDevice)

	port := config.GetPort()
	log.Fatal(app.Listen(":" + port))
}
