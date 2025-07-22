package config

import (
	"log"
	"os"
	"strconv"
)

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	if port, err := strconv.Atoi(port); err != nil || port < 1 || port > 65535 {
		log.Fatalf("Invalid port number: %d", port)
	}

	return port
}

func GetAPIKey() string {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY is not set")
	}
	return apiKey
}

func GetDatabaseURL() string {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/dmt_db?sslmode=disable"
	}
	return dbURL
}
