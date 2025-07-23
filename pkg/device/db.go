package device

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
)

func InsertDevice(ctx context.Context, db *pgx.Conn, device *Device) error {
	sanitizeDevice(device)

	if validationErrors := validateDevice(device); len(validationErrors) > 0 {
		message := ""
		for _, err := range validationErrors {
			message += err.Error() + "; "
		}

		return fmt.Errorf("validation failed: %s", message)
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

// Only updates employee field for now
func UpdateDevice(ctx context.Context, db *pgx.Conn, device *Device) error {
	if device.ID < 1 {
		return errors.New("device ID is required")
	}

	args := []interface{}{}
	sqlChunk := []string{}

	if device.Employee != nil {
		employee := strings.TrimSpace(*device.Employee)
		if employee != "" {
			args = append(args, employee)
			sqlChunk = append(sqlChunk, fmt.Sprintf(" employee = $%d", len(args)))
		} else {
			sqlChunk = append(sqlChunk, " employee = NULL")
		}
	}

	if len(sqlChunk) == 0 {
		return errors.New("no update options provided")
	}

	strBuilder := strings.Builder{}
	strBuilder.WriteString("UPDATE device SET")

	for i, chunk := range sqlChunk {
		strBuilder.WriteString(chunk)

		if i < len(sqlChunk)-1 {
			strBuilder.WriteString(",")
		}
	}

	fmt.Fprintf(&strBuilder, " WHERE id = %d RETURNING id, created_at, updated_at, name, type, ip, mac, description, employee", device.ID)

	query := strBuilder.String()

	log.Infof("Executing query: %s", query)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := db.QueryRow(ctx, query, args...).Scan(
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

func GetDevices(ctx context.Context, db *pgx.Conn, employee, deviceType, ip, mac string) ([]Device, error) {
	query := `
		SELECT id, created_at, updated_at, name, type, ip, mac, description, employee 
		FROM device 
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if employee != "" {
		query += fmt.Sprintf(" AND employee = $%d", argIndex)
		args = append(args, employee)
		argIndex++
	}

	if deviceType != "" {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, deviceType)
		argIndex++
	}

	if ip != "" {
		query += fmt.Sprintf(" AND ip LIKE $%d", argIndex)
		args = append(args, "%"+ip+"%")
		argIndex++
	}

	if mac != "" {
		query += fmt.Sprintf(" AND mac ILIKE $%d", argIndex)
		args = append(args, "%"+mac+"%")
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	devices, err := pgx.CollectRows(rows, pgx.RowToStructByName[Device])

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return devices, nil
}
