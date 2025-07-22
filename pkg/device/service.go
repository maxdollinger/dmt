package device

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type DeviceService struct {
	db *pgx.Conn
}

func NewDeviceService(db *pgx.Conn) *DeviceService {
	return &DeviceService{db: db}
}

type CreateDeviceRequest struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	IP          string  `json:"ip"`
	MAC         string  `json:"mac"`
	Description *string `json:"description,omitempty"`
	Employee    *string `json:"employee,omitempty"`
}

func (s *DeviceService) CreateDevice(c *fiber.Ctx) error {
	device := new(Device)
	// Parse JSON body
	err := c.BodyParser(&device)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	err = device.Insert(c.Context(), s.db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create device",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Device created successfully",
		"device":  device,
	})
}
