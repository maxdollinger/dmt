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
