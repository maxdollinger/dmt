package internals

import (
	"dmt/internals/middleware"
	"dmt/pkg/device"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5"
)

func CreateApp(db *pgx.Conn, apiKey string) *fiber.App {
	app := fiber.New()

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middleware.KeyAuthMiddleware(apiKey))

	deviceService := device.NewDeviceService(db)

	app.Post("/devices", deviceService.CreateDevice)
	app.Get("/devices/:id", deviceService.GetDevice)
	app.Delete("/devices/:id", deviceService.DeleteDevice)

	return app
}
