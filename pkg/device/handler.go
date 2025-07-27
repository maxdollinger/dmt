package device

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeviceHandler struct {
	db *pgxpool.Pool
}

func NewDeviceHandler(db *pgxpool.Pool) *DeviceHandler {
	return &DeviceHandler{db: db}
}

func (s *DeviceHandler) CreateDevice(c *fiber.Ctx) error {
	device := new(Device)
	err := c.BodyParser(&device)
	if err != nil {
		log.Errorf("Failed to parse device: %s", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	err = InsertDevice(c.Context(), s.db, device)
	if err != nil {
		log.Errorf("Failed to create device: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create device",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Device created successfully",
		"device":  device,
	})
}

func (s *DeviceHandler) GetDeviceByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		log.Errorf("Invalid device ID: %s", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid device ID",
		})
	}

	device := &Device{ID: id}
	err = GetDeviceByID(c.Context(), s.db, device)
	if err != nil {
		log.Errorf("Failed to retrieve device: %s", err.Error())
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

func (s *DeviceHandler) DeleteDevice(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		log.Errorf("Invalid device ID: %s", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid device ID",
		})
	}

	device := &Device{ID: id}
	err = DeleteDevice(c.Context(), s.db, device)
	if err != nil {
		log.Errorf("Failed to delete device: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete device",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Device deleted successfully",
	})
}

func (s *DeviceHandler) UpdateDeviceEmployee(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		log.Errorf("Invalid device ID: %s", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid device ID",
		})
	}

	var requestBody struct {
		Employee string `json:"employee"`
	}
	err = c.BodyParser(&requestBody)
	if err != nil {
		log.Errorf("Invalid JSON format: %s", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	device := &Device{ID: id, Employee: &requestBody.Employee}
	err = UpdateDevice(c.Context(), s.db, device)
	if err != nil {
		log.Errorf("Failed to update device: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update device employee",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Device employee updated successfully",
		"device":  device,
	})
}

func (s *DeviceHandler) DeleteDeviceEmployee(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		log.Errorf("Invalid device ID: %s", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid device ID",
		})
	}

	employee := ""
	device := &Device{ID: id, Employee: &employee}
	err = UpdateDevice(c.Context(), s.db, device)
	if err != nil {
		log.Errorf("Failed to remove device employee: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove device employee",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Device employee removed successfully",
		"device":  device,
	})
}

func (s *DeviceHandler) GetDevices(c *fiber.Ctx) error {
	employee := c.Query("employee")
	deviceType := c.Query("type")
	ip := c.Query("ip")
	mac := c.Query("mac")

	devices, err := GetDevices(c.Context(), s.db, employee, deviceType, ip, mac)
	if err != nil {
		log.Errorf("Failed to retrieve devices: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve devices",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"devices": devices,
		"count":   len(devices),
	})
}
