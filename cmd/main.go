package main

import (
	"gomatrixbot/internal/gomatrixbot"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Set DB file
	gomatrixbot.DBFile = "gomatrixbot.db"
	dbFile := os.Getenv("SQLITE3_DB")
	if dbFile != "" {
		gomatrixbot.DBFile = dbFile
	}

	// Set Matrix Credentials
	gomatrixbot.MatrixHost = os.Getenv("MATRIX_HOST")
	gomatrixbot.MatrixUsername = os.Getenv("MATRIX_USERNAME")
	gomatrixbot.MatrixPassword = os.Getenv("MATRIX_PASSWORD")

	// Start application
	gomatrixbot.Run()
}
