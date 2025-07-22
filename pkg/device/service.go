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
	err := c.BodyParser(&device)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	err = InsertDevice(c.Context(), s.db, device)
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
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid device ID",
		})
	}

	device := &Device{ID: id}
	err = GetDeviceByID(c.Context(), s.db, device)
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

func (s *DeviceService) DeleteDevice(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid device ID",
		})
	}

	device := &Device{ID: id}
	err = DeleteDevice(c.Context(), s.db, device)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete device",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Device deleted successfully",
	})
}

func (s *DeviceService) UpdateDevice(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid device ID",
		})
	}

	device := &Device{ID: id}
	err = c.BodyParser(&device)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	err = UpdateDevice(c.Context(), s.db, device)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update device",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Device updated successfully",
		"device":  device,
	})
}

func (s *DeviceService) GetDevices(c *fiber.Ctx) error {
	employee := c.Query("employee")
	deviceType := c.Query("type")
	ip := c.Query("ip")
	mac := c.Query("mac")

	devices, err := GetDevices(c.Context(), s.db, employee, deviceType, ip, mac)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve devices",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"devices": devices,
		"count":   len(devices),
	})
}
