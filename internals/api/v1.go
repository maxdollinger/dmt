package api

import (
	"dmt/pkg/device"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func CreateV1Router(db *pgx.Conn) *fiber.App {
	v1 := fiber.New()

	deviceService := device.NewDeviceService(db)

	// Device routes
	v1.Post("/devices", deviceService.CreateDevice)
	v1.Get("/devices", deviceService.GetDevices)
	v1.Get("/devices/:id", deviceService.GetDevice)
	v1.Put("/devices/:id/employee", deviceService.UpdateDeviceEmployee)
	v1.Delete("/devices/:id/employee", deviceService.DeleteDeviceEmployee)
	v1.Delete("/devices/:id", deviceService.DeleteDevice)

	return v1
}
