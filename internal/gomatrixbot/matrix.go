package gomatrixbot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
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

	ducks       duckHunt
	quotes      map[id.RoomID]quoteCache
	roomAliases map[id.RoomID]string
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
		go mtrx.handleEvent(source, evt)
	})

	syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		go mtrx.handleInvite(source, evt)
	})

	// Init all the things
	go mtrx.initDuckHunt()
	go mtrx.initQuote()
	mtrx.roomAliases = make(map[id.RoomID]string)

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
		mtrx.parseCommand(source, evt)
		return
	}

	mtrx.appendLastMessage(evt.RoomID, evt.Sender, evt.Content.AsMessage().Body)
}

func (mtrx *MtrxClient) handleInvite(source mautrix.EventSource, evt *event.Event) {
	// If syncing older data.. stop that
	if evt.Timestamp < mtrx.startTime {
		return
	}

	content := struct {
		Inviter id.UserID `json:"inviter"`
	}{evt.Sender}

	if _, err := mtrx.c.JoinRoom(evt.RoomID.String(), "", content); err != nil {
		fmt.Printf("Failed to join room: %s", evt.RoomID.String())
	} else {
		fmt.Printf("Joined room: %s", evt.RoomID.String())
	}
}

func (mtrx *MtrxClient) parseCommand(source mautrix.EventSource, evt *event.Event) {
	cmd := strings.Split(evt.Content.AsMessage().Body[1:], " ")
	if len(cmd) == 0 {
		return
	}

	switch cmd[0] {
	case "help":
		msg := "@dbot:mtrx.nz commands:\n\n" +
			"!echo - echo message back to channel\n" +
			"!quote <user>, !quote - quote users last message or returns a random quote\n" +
			"!starthunt, !stophunt, !bang - duckhunt commands"
		_, err := mtrx.c.SendNotice(evt.RoomID, msg)
		if err != nil {
			log.Print(err)
		}
	case "echo":
		_, err := mtrx.c.SendNotice(
			evt.RoomID, strings.Join(cmd[1:], " "))
		if err != nil {
			log.Print(err)
		}
	case "quote":
		if len(cmd) == 1 {
			user, quote := mtrx.getRandomQuote(evt.RoomID)
			_, err := mtrx.c.SendNotice(evt.RoomID, fmt.Sprintf("Quote from %s: %s", user, quote))
			if err != nil {
				log.Print(err)
			}
			return
		} else if len(cmd) == 2 {
			mtrx.quote(evt.RoomID, id.UserID(cmd[1]))
		} else {
			_, err := mtrx.c.SendNotice(evt.RoomID, "Usage - !quote <user>")
			if err != nil {
				log.Print(err)
			}
			return
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
