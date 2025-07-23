package device

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
)

type Notification struct {
	Employee string `json:"employee"`
	Count    int    `json:"count"`
}

// NotifyDeviceCount listens for PostgreSQL notifications on the device_count channel
// and returns a channel that emits parsed Notification structs.
// The function will run until the context is cancelled or an unrecoverable error occurs.
func NotifyDeviceCount(ctx context.Context, url string) <-chan Notification {
	notificationChan := make(chan Notification, 10) // Buffered to avoid blocking

	go func() {
		defer close(notificationChan)

		conn, err := pgx.Connect(ctx, url)
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %s", err)
		}
		defer conn.Close(ctx)

		err = conn.Ping(ctx)
		if err != nil {
			log.Fatalf("Failed to ping PostgreSQL: %s", err)
		}

		_, err = conn.Exec(ctx, "LISTEN device_count")
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
				// Set a timeout for waiting for notifications to allow context checking
				notifyCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
				pgNotification, err := conn.WaitForNotification(notifyCtx)
				cancel()

				if err != nil {
					if ctx.Err() != nil {
						return
					}
					continue
				}

				var deviceNotification Notification
				if err := json.Unmarshal([]byte(pgNotification.Payload), &deviceNotification); err != nil {
					log.Errorf("Failed to parse notification payload: %v", err)
					continue
				}

				log.Infof("Received device count notification - Employee: %s, Count: %d", deviceNotification.Employee, deviceNotification.Count)

				notificationChan <- deviceNotification
			}
		}
	}()

	return notificationChan
}
