package main

import (
	"gomatrixbot/internal/gomatrixbot"
	"os"
)

func main() {
	// Set Matrix Credentials
	gomatrixbot.MatrixHost = os.Getenv("MATRIX_HOST")
	gomatrixbot.MatrixUsername = os.Getenv("MATRIX_USERNAME")
	gomatrixbot.MatrixPassword = os.Getenv("MATRIX_PASSWORD")

	// Debuggerrr. Shitty env work around :ok_hand:
	debugFlag := os.Getenv("DEBUG")
	gomatrixbot.Debug = false
	if debugFlag == "1" || debugFlag == "true" {
		gomatrixbot.Debug = true
	}

	gomatrixbot.Admin = os.Getenv("ADMIN_USER")

	// Start application
	gomatrixbot.Run()
}
