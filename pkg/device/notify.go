package device

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	Employee string `json:"employee"`
	Count    int    `json:"count"`
}

type NotificationRequest struct {
	Level                string `json:"level"`
	EmployeeAbbreviation string `json:"employeeAbbreviation"`
	Message              string `json:"message"`
}

func HandleDeviceCountNotifications(ctx context.Context, db *pgxpool.Pool, notificationUrl string) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection from pool: %w", err)
	}

	notificationChan, err := DeviceCountListener(ctx, conn.Conn())
	if err != nil {
		conn.Release()
		return fmt.Errorf("failed to start device count listener: %w", err)
	}

	go func() {
		defer conn.Release()
		defer log.Info("Notification handler stopped")

		for notification := range notificationChan {
			if notificationUrl != "" {
				SendNotification(notificationUrl, &notification)
			} else {
				log.Warnf("Notification URL is not set, skipping notification")
			}
		}
	}()

	return nil
}

func SendNotification(notificationUrl string, notification *Notification) {
	notificationReq := NotificationRequest{
		Level:                "warning",
		EmployeeAbbreviation: notification.Employee,
		Message:              fmt.Sprintf("Device count warning: Employee %s has %d devices", notification.Employee, notification.Count),
	}

	jsonData, err := json.Marshal(notificationReq)
	if err != nil {
		log.Errorf("failed to marshal notification request: %v", err)
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(notificationUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Errorf("failed to send notification: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Errorf("notification service returned status code: %d", resp.StatusCode)
		return
	}

	log.Infof("Successfully sent notification - Message: %s", notificationReq.Message)
}

func DeviceCountListener(ctx context.Context, conn *pgx.Conn) (<-chan Notification, error) {
	notificationChan := make(chan Notification, 10)

	_, err := conn.Exec(ctx, "LISTEN device_count")
	if err != nil {
		close(notificationChan)
		return nil, fmt.Errorf("failed to listen for device count notifications: %w", err)
	}

	go func() {
		defer close(notificationChan)
		defer log.Info("Device count listener stopped")

		log.Info("Started listening for device count notifications")

		for {
			select {
			case <-ctx.Done():
				log.Info("Stopping device count notification listener")
				return
			default:
				pgNotification, err := conn.WaitForNotification(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					log.Errorf("Failed to wait for notification: %v", err)
					continue
				}

				var deviceNotification Notification
				if err := json.Unmarshal([]byte(pgNotification.Payload), &deviceNotification); err != nil {
					log.Errorf("Failed to parse notification payload: %v", err)
					continue
				}

				log.Infof("Received device count notification - Employee: %s, Count: %d", deviceNotification.Employee, deviceNotification.Count)

				if deviceNotification.Count >= 3 {
					select {
					case notificationChan <- deviceNotification:
						log.Infof("Notification sent to alarm service")
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return notificationChan, nil
}
