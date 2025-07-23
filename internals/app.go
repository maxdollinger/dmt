package internals

import (
	"dmt/internals/api"
	"dmt/internals/middleware"

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

	// Mount v1 API routes
	v1Router := api.CreateV1Router(db)
	app.Mount("/api/v1", v1Router)

	return app
}
