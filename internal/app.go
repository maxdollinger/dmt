package internal

import (
	"dmt/internal/middleware"
	"dmt/pkg/device"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateHttpServer(db *pgxpool.Pool, apiKey string) *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit: 512,
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(healthcheck.New())

	api := app.Group("/api")
	api.Use(middleware.KeyAuthMiddleware(apiKey))

	v1 := api.Group("/v1")

	deviceHandler := device.NewDeviceHandler(db)

	v1.Post("/devices", deviceHandler.CreateDevice)
	v1.Get("/devices", deviceHandler.GetDevices)
	v1.Get("/devices/:id", deviceHandler.GetDeviceByID)
	v1.Delete("/devices/:id", deviceHandler.DeleteDevice)
	v1.Put("/devices/:id/employee", deviceHandler.UpdateDeviceEmployee)
	v1.Delete("/devices/:id/employee", deviceHandler.DeleteDeviceEmployee)

	return app
}
