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
	defer db.Close()

	device.HandleDeviceCountNotifications(ctx, db, notificationUrl)

	server := internals.CreateHttpServer(db, apiKey)

	port := config.GetPort()
	log.Fatal(server.Listen(":" + port))
}
