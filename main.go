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
	notificationUrl := config.GetNotifyUrl()

	ctx := context.Background()
	db := internals.ConnectDb(ctx, databaseURL)
	defer db.Close(ctx)

	notificationChan := device.DeviceCountListener(ctx, databaseURL)
	go func() {
		for notification := range notificationChan {
			device.SendNotification(notificationUrl, &notification)
		}
	}()

	app := internals.CreateApp(db, apiKey)

	port := config.GetPort()
	log.Fatal(app.Listen(":" + port))
}
