package main

import (
	"dmt/config"
	"dmt/internals/middleware"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	apiKey := config.GetAPIKey()

	app := fiber.New()

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middleware.KeyAuthMiddleware(apiKey))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	port := config.GetPort()
	log.Fatal(app.Listen(":" + port))
}
