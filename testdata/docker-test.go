package main

import (
	"fmt"
	"os"
)

func main() {
	// Variables from Dockerfile
	apiKey := os.Getenv("API_KEY")
	dbUrl := os.Getenv("DATABASE_URL")

	// Variable from .env
	maxRetries := os.Getenv("MAX_RETRIES")

	fmt.Printf("API Key: %s\n", apiKey)
	fmt.Printf("Database: %s\n", dbUrl)
	fmt.Printf("Max Retries: %s\n", maxRetries)
}
