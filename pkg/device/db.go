package device

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func InsertDevice(ctx context.Context, db *pgx.Conn, device *Device) error {
	sanitizeDevice(device)

	if validationErrors := validateDevice(device); len(validationErrors) > 0 {
		message := ""
		for _, err := range validationErrors {
			message += err.Error() + "; "
		}

		return fmt.Errorf("Validation failed: %s", message)
	}

	query := `
		INSERT INTO device (name, type, ip, mac, description, employee) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := db.QueryRow(ctx, query,
		device.Name,
		device.Type,
		device.IP,
		device.MAC,
		device.Description,
		device.Employee,
	).Scan(&device.ID, &device.CreatedAt, &device.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func UpdateDevice(ctx context.Context, db *pgx.Conn, device *Device) error {
	sanitizeDevice(device)

	if validationErrors := validateDevice(device); len(validationErrors) > 0 {
		message := ""
		for _, err := range validationErrors {
			message += err.Error() + "; "
		}

		return fmt.Errorf("Validation failed: %s", message)
	}

	query := `
		UPDATE device 
		SET name = $1, type = $2, ip = $3, mac = $4, description = $5, employee = $6 
		WHERE id = $7 
		RETURNING updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := db.QueryRow(ctx, query,
		device.Name,
		device.Type,
		device.IP,
		device.MAC,
		device.Description,
		device.Employee,
		device.ID,
	).Scan(&device.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func DeleteDevice(ctx context.Context, db *pgx.Conn, device *Device) error {
	query := `
		DELETE FROM device 
		WHERE id = $1 
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.Exec(ctx, query, device.ID)
	if err != nil {
		return err
	}

	return nil
}

func GetDeviceByID(ctx context.Context, db *pgx.Conn, device *Device) error {
	query := `
		SELECT id, created_at, updated_at, name, type, ip, mac, description, employee 
		FROM device 
		WHERE id = $1 
		LIMIT 1
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := db.QueryRow(ctx, query, device.ID).Scan(
		&device.ID,
		&device.CreatedAt,
		&device.UpdatedAt,
		&device.Name,
		&device.Type,
		&device.IP,
		&device.MAC,
		&device.Description,
		&device.Employee,
	)
	if err != nil {
		return err
	}

	return nil
}
