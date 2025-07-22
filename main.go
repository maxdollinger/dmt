package main

import (
	"context"
	"dmt/config"
	"dmt/internals"
	"log"
)

func main() {
	apiKey := config.GetAPIKey()
	databaseURL := config.GetDatabaseURL()

	ctx := context.Background()
	db := internals.ConnectDb(ctx, databaseURL)
	defer db.Close(ctx)

	app := internals.CreateApp(db, apiKey)

	port := config.GetPort()
	log.Fatal(app.Listen(":" + port))
}
