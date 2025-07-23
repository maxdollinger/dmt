package api

import (
	"dmt/pkg/device"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func CreateV1Router(db *pgx.Conn) *fiber.App {
	v1Router := fiber.New()

	deviceService := device.NewDeviceService(db)

	v1Router.Post("/devices", deviceService.CreateDevice)
	v1Router.Get("/devices", deviceService.GetDevices)
	v1Router.Get("/devices/:id", deviceService.GetDevice)
	v1Router.Put("/devices/:id/employee", deviceService.UpdateDeviceEmployee)
	v1Router.Delete("/devices/:id/employee", deviceService.DeleteDeviceEmployee)
	v1Router.Delete("/devices/:id", deviceService.DeleteDevice)

	return v1Router
}
