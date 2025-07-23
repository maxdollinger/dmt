package main

import (
	"context"
	"dmt/internals"
	"dmt/internals/config"
	"dmt/pkg/device"
	"log"
)

func main() {
	apiKey := config.GetAPIKey()
	databaseURL := config.GetDatabaseURL()

	ctx := context.Background()
	db := internals.ConnectDb(ctx, databaseURL)
	defer db.Close(ctx)

	notificationChan := device.NotifyDeviceCount(ctx, databaseURL)
	go func() {
		for notification := range notificationChan {
			log.Printf("Device count alert: Employee %s has %d devices",
				notification.Employee, notification.Count)
		}
	}()

	app := internals.CreateApp(db, apiKey)

	port := config.GetPort()
	log.Fatal(app.Listen(":" + port))
}
