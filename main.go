package main

import (
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New())

	port := getPort()
	log.Fatal(app.Listen(":" + port))
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	if port, err := strconv.Atoi(port); err != nil || port < 1 || port > 65535 {
		log.Fatalf("Invalid port number: %d", port)
	}

	return port
}
