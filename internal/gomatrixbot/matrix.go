package gomatrixbot

import (
	"fmt"
	"log"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

var (
	// Matrix creds
	MatrixHost     string
	MatrixUsername string
	MatrixToken    string

	// Mautrix client
	c *mautrix.Client
)

// Run - starts bot!
func Run() {
	c = initClient()

	log.Printf("Started gomatrixbot")

	roomResp, err := c.JoinRoom("#testroom:mtrx.nz", "", nil)
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.SendNotice(roomResp.RoomID, "test!")
	if err != nil {
		log.Print(err)
	}

	// Start syncing stuff
	// err := matrixClient.Sync()
	if err != nil {
		log.Print(err)
	}
}

func initClient() *mautrix.Client {
	mautrix.DefaultUserAgent = fmt.Sprintf("gomatrixbot %s", mautrix.DefaultUserAgent)
	userID := id.UserID(MatrixUsername)

	// Connect new client
	client, err := mautrix.NewClient(MatrixHost, userID, MatrixToken)
	if err != nil {
		log.Fatal(err)
	}

	return client
}
