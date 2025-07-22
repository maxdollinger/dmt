package device

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (d *Device) Insert(ctx context.Context, db *pgx.Conn) error {
	sanitizeDevice(d)

	if validationErrors := validateDevice(d); len(validationErrors) > 0 {
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
		d.Name,
		d.Type,
		d.IP,
		d.MAC,
		d.Description,
		d.Employee,
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (d *Device) Update(ctx context.Context, db *pgx.Conn) error {
	sanitizeDevice(d)

	if validationErrors := validateDevice(d); len(validationErrors) > 0 {
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
		d.Name,
		d.Type,
		d.IP,
		d.MAC,
		d.Description,
		d.Employee,
		d.ID,
	).Scan(&d.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (d *Device) Delete(ctx context.Context, db *pgx.Conn) error {
	query := `
		DELETE FROM device 
		WHERE id = $1 
		RETURNING updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := db.QueryRow(ctx, query, d.ID).Scan(&d.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (d *Device) GetByID(ctx context.Context, db *pgx.Conn) error {
	query := `
		SELECT id, created_at, updated_at, name, type, ip, mac, description, employee 
		FROM device 
		WHERE id = $1 
		LIMIT 1
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := db.QueryRow(ctx, query, d.ID).Scan(
		&d.ID,
		&d.CreatedAt,
		&d.UpdatedAt,
		&d.Name,
		&d.Type,
		&d.IP,
		&d.MAC,
		&d.Description,
		&d.Employee,
	)
	if err != nil {
		return err
	}

	return nil
}
