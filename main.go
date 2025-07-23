package main

import (
	"context"
	"dmt/internal"
	"dmt/internal/config"
	"dmt/pkg/device"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	apiKey := config.GetAPIKey()
	databaseURL := config.GetDatabaseURL()
	notificationUrl := config.GetNotifyUrl()
	port := config.GetPort()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := internal.ConnectDb(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := device.HandleDeviceCountNotifications(ctx, db, notificationUrl); err != nil {
		log.Fatalf("Failed to start notification handler: %v", err)
	}

	server := internal.CreateHttpServer(db, apiKey)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Listen(":" + port); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-quit
	cancel()
	_ = server.Shutdown()
	log.Println("Server shutdown complete")

	db.Close()
	log.Println("Database connection closed")

	log.Println("Application shutdown complete")
}
