package device

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type DeviceService struct {
	db *pgx.Conn
}

func NewDeviceService(db *pgx.Conn) *DeviceService {
	return &DeviceService{db: db}
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

func (s *DeviceService) GetDevice(c *fiber.Ctx) error {
	// Parse ID from URL parameter
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid device ID",
		})
	}

	// Create device instance with ID and fetch from database
	device := &Device{ID: id}
	err = device.GetByID(c.Context(), s.db)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Device not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve device",
		})
	}

	return c.Status(fiber.StatusOK).JSON(device)
}
