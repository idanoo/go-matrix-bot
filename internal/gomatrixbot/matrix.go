package gomatrixbot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

var (
	// Matrix creds
	MatrixHost     string
	MatrixUsername string
	MatrixPassword string
)

type MtrxClient struct {
	c         *mautrix.Client
	startTime int64
	db        *sql.DB
}

// Run - starts bot!
func Run() {
	mtrx := MtrxClient{}

	// Get DB connection and run any pending migrations
	mtrx.db = InitDb()
	defer mtrx.CloseDbConn()

	// Connet to matrix!
	mtrx.c = initClient()
	mtrx.startTime = time.Now().UnixMilli()

	log.Printf("Started gomatrixbot")

	syncer := mtrx.c.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
		mtrx.handleEvent(source, evt)
	})

	// Launch'er up
	err := mtrx.c.Sync()
	if err != nil {
		log.Fatal(err)
	}
}

func initClient() *mautrix.Client {
	// Connect new client
	client, err := mautrix.NewClient(MatrixHost, "", "")
	if err != nil {
		log.Fatal(err)
	}

	// Login
	_, err = client.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: MatrixUsername},
		Password:         MatrixPassword,
		StoreCredentials: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func (mtrx *MtrxClient) handleEvent(source mautrix.EventSource, evt *event.Event) {
	// If syncing older messages.. stop that
	if evt.Timestamp < mtrx.startTime {
		return
	}

	// If parsing own messages.. stop that too
	if evt.Sender == mtrx.c.UserID {
		return
	}

	// Parse command if intended for us
	if evt.Content.AsMessage().Body[0:1] == "!" {
		fmt.Printf("<%[1]s> %[4]s (%[2]s/%[3]s)\n", evt.Sender, evt.Type.String(), evt.ID, evt.Content.AsMessage().Body)
		go mtrx.parseCommand(source, evt)
		return
	}
}

func (mtrx *MtrxClient) parseCommand(source mautrix.EventSource, evt *event.Event) {
	cmd := strings.Split(evt.Content.AsMessage().Body[1:], " ")
	if len(cmd) == 0 {
		return
	}

	switch cmd[0] {
	case "help":
		msg := "gomatrixbot commands:\n\n" +
			"!echo - echo message back to channel\n" +
			"!starthunt, !stophunt, !bang - duckhunt commands"
		_, err := mtrx.c.SendNotice(evt.RoomID, msg)
		if err != nil {
			log.Print(err)
		}
	case "echo":
		_, err := mtrx.c.SendNotice(evt.RoomID, strings.Join(cmd[1:], " "))
		if err != nil {
			log.Print(err)
		}
	case "starthunt":
		fallthrough
	case "stophunt":
		fallthrough
	case "bang":
		_, err := mtrx.c.SendNotice(evt.RoomID, "Duckhunt coming soon...")
		if err != nil {
			log.Print(err)
		}
	}
}
