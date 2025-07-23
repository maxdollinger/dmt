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

func HandleDeviceCountNotifications(ctx context.Context, db *pgxpool.Pool, notificationUrl string) {
	conn, err := db.Acquire(ctx)
	if err != nil {
		log.Fatalf("Failed to acquire connection from pool: %s", err.Error())
	}

	notificationChan := DeviceCountListener(ctx, conn.Conn())
	go func() {
		for notification := range notificationChan {
			SendNotification(notificationUrl, &notification)
		}
	}()
}

func SendNotification(notificationUrl string, notification *Notification) {
	notificationReq := NotificationRequest{
		Level:                "warning",
		EmployeeAbbreviation: notification.Employee,
		Message:              fmt.Sprintf("Device count warning: Employee %s has %d devices", notification.Employee, notification.Count),
	}

	jsonData, err := json.Marshal(notificationReq)
	if err != nil {
		log.Errorf("failed to marshal notification request: %w", err)
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(notificationUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Errorf("failed to send notification: %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Errorf("notification service returned status code: %d", resp.StatusCode)
		return
	}

	log.Infof("Successfully sent notification - Message: %s", notificationReq.Message)
}

func DeviceCountListener(ctx context.Context, conn *pgx.Conn) <-chan Notification {
	notificationChan := make(chan Notification, 10)

	go func() {
		defer close(notificationChan)

		_, err := conn.Exec(ctx, "LISTEN device_count")
		if err != nil {
			log.Fatalf("Failed to listen for device count notifications: %s", err)
		}

		log.Info("Started listening for device count notifications")

		for {
			select {
			case <-ctx.Done():
				log.Info("Stopping device count notification listener")
				return
			default:
				pgNotification, err := conn.WaitForNotification(ctx)
				if err != nil {
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
					notificationChan <- deviceNotification
				}
			}
		}
	}()

	return notificationChan
}
