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
	mtrx MtrxClient
)

type MtrxClient struct {
	c     *mautrix.Client
	rooms map[string]string
}

// Run - starts bot!
func Run() {
	mtrx := MtrxClient{}
	mtrx.c = initClient()
	mtrx.rooms = db.getRooms()

	// Join previously invited rooms.
	for _, roomName := range mtrx.rooms {
		_, err := mtrx.c.JoinRoom(roomName, "", nil)
		if err != nil {
			log.Print(err)
		}
	}

	log.Printf("Started gomatrixbot")

	roomResp, err := mtrx.c.JoinRoom("#testroom:mtrx.nz", "", nil)
	if err != nil {
		log.Fatal(err)
	}

	_, err = mtrx.c.SendNotice(roomResp.RoomID, "test!")
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
