package gomatrixbot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/chzyer/readline"
	"github.com/rs/zerolog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var (
	// Matrix creds
	MatrixHost     string
	MatrixUsername string
	MatrixPassword string
	Debug          bool

	BotDb     = "/config/bot.db"
	MautrixDB = "/config/mautrix.db"
)

type MtrxClient struct {
	c           *mautrix.Client
	startTime   int64
	db          *sql.DB
	olm         *crypto.OlmMachine
	cryptoStore *crypto.SQLCryptoStore

	// ducks         duckHunt
	// quotes        map[id.RoomID]quoteCache
	// roomEncrypted map[id.RoomID]bool
}

// Run - starts bot!
func Run() {
	mtrx := MtrxClient{}

	// Get DB connection and run any pending migrations
	mtrx.db = InitDb()
	defer mtrx.CloseDbConn()

	client, err := mautrix.NewClient(MatrixHost, "", "")
	if err != nil {
		panic(err)
	}
	rl, err := readline.New("[no room]> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()
	log := zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = rl.Stdout()
		w.TimeFormat = time.Stamp
	})).With().Timestamp().Logger()
	if !Debug {
		log = log.Level(zerolog.InfoLevel)
	}
	client.Log = log

	var lastRoomID id.RoomID

	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
		lastRoomID = evt.RoomID
		rl.SetPrompt(fmt.Sprintf("%s> ", lastRoomID))
		log.Info().
			Str("sender", evt.Sender.String()).
			Str("type", evt.Type.String()).
			Str("id", evt.ID.String()).
			Str("body", evt.Content.AsMessage().Body).
			Msg("Received message")
	})
	syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		if evt.GetStateKey() == client.UserID.String() && evt.Content.AsMember().Membership == event.MembershipInvite {
			_, err := client.JoinRoomByID(evt.RoomID)
			if err == nil {
				lastRoomID = evt.RoomID
				rl.SetPrompt(fmt.Sprintf("%s> ", lastRoomID))
				log.Info().
					Str("room_id", evt.RoomID.String()).
					Str("inviter", evt.Sender.String()).
					Msg("Joined room after invite")
			} else {
				log.Error().Err(err).
					Str("room_id", evt.RoomID.String()).
					Str("inviter", evt.Sender.String()).
					Msg("Failed to join room after invite")
			}
		}
	})

	cryptoHelper, err := cryptohelper.NewCryptoHelper(client, []byte("meow"), MautrixDB)
	if err != nil {
		panic(err)
	}

	// You can also store the user/device IDs and access token and put them in the client beforehand instead of using LoginAs.
	//client.UserID = "..."
	//client.DeviceID = "..."
	//client.AccessToken = "..."
	// You don't need to set a device ID in LoginAs because the crypto helper will set it for you if necessary.
	cryptoHelper.LoginAs = &mautrix.ReqLogin{
		Type:       mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: MatrixUsername},
		Password:   MatrixPassword,
	}
	// If you want to use multiple clients with the same DB, you should set a distinct database account ID for each one.
	//cryptoHelper.DBAccountID = ""
	err = cryptoHelper.Init()
	if err != nil {
		panic(err)
	}
	// Set the client crypto helper in order to automatically encrypt outgoing messages
	client.Crypto = cryptoHelper

	log.Info().Msg("Now running")
	syncCtx, cancelSync := context.WithCancel(context.Background())
	var syncStopWait sync.WaitGroup
	syncStopWait.Add(1)

	go func() {
		err = client.SyncWithContext(syncCtx)
		defer syncStopWait.Done()
		if err != nil && !errors.Is(err, context.Canceled) {
			panic(err)
		}
	}()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		if lastRoomID == "" {
			log.Error().Msg("Wait for an incoming message before sending messages")
			continue
		}
		resp, err := client.SendText(lastRoomID, line)
		if err != nil {
			log.Error().Err(err).Msg("Failed to send event")
		} else {
			log.Info().Str("event_id", resp.EventID.String()).Msg("Event sent")
		}
	}
	cancelSync()
	syncStopWait.Wait()
	err = cryptoHelper.Close()
	if err != nil {
		log.Error().Err(err).Msg("Error closing database")
	}

	log.Printf("Started gomatrixbot")

	// Init all the things
	// go mtrx.initDuckHunt()
	// go mtrx.initQuote()
	// go mtrx.initRss()

	// // Launch'er up
	// err := mtrx.c.Sync()
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

// func (mtrx *MtrxClient) handleEvent(source mautrix.EventSource, evt *event.Event) {
// 	// If syncing older messages.. stop that
// 	if evt.Timestamp < mtrx.startTime {
// 		return
// 	}

// 	// If parsing own messages.. stop that too
// 	if evt.Sender == mtrx.c.UserID {
// 		return
// 	}

// 	// Parse command if intended for us
// 	if evt.Content.AsMessage().Body[0:1] == "!" {
// 		fmt.Printf("<%[1]s> %[4]s (%[2]s/%[3]s)\n", evt.Sender, evt.Type.String(), evt.ID, evt.Content.AsMessage().Body)
// 		mtrx.parseCommand(source, evt)
// 		return
// 	}

// 	mtrx.appendLastMessage(evt.RoomID, evt.Sender, evt.Content.AsMessage().Body)
// }

// func (mtrx *MtrxClient) handleInvite(source mautrix.EventSource, evt *event.Event) {
// 	// If syncing older data.. stop that
// 	if evt.Timestamp < mtrx.startTime {
// 		return
// 	}

// 	if evt.Sender == mtrx.c.UserID {
// 		return
// 	}

// 	content := struct {
// 		Inviter id.UserID `json:"inviter"`
// 	}{evt.Sender}

// 	fmt.Println("Invite by" + evt.Sender.String())

// 	if _, err := mtrx.c.JoinRoom(evt.RoomID.String(), "", content); err != nil {
// 		fmt.Printf("Failed to join room: %s", evt.RoomID.String())
// 	} else {
// 		fmt.Printf("Joined room: %s", evt.RoomID.String())
// 	}
// }

// func (mtrx *MtrxClient) parseCommand(source mautrix.EventSource, evt *event.Event) {
// 	cmd := strings.Split(strings.TrimSpace(evt.Content.AsMessage().Body[1:]), " ")
// 	if len(cmd) == 0 {
// 		return
// 	}

// 	switch cmd[0] {
// 	case "help":
// 		msg := "@dbot:mtrx.nz commands:\n" +
// 			"\n!echo - echo message back to channel" +
// 			"\n!quote <user>, !quote - quote users last message or returns a random quote" +
// 			"\n!starthunt, !stophunt, !bang - duckhunt commands" +
// 			"\n!rss - RSS subscription commands"
// 		mtrx.sendMessage(evt.RoomID, msg)
// 	case "echo":
// 		mtrx.sendMessage(evt.RoomID, strings.Join(cmd[1:], " "))
// 	case "quote":
// 		if len(cmd) == 1 {
// 			user, quote := mtrx.getRandomQuote(evt.RoomID)
// 			mtrx.sendMessage(evt.RoomID, fmt.Sprintf("Quote from %s: %s", user, quote))
// 			return
// 		} else if len(cmd) == 2 {
// 			re := regexp.MustCompile(`@.*\.[^"]+`)
// 			match := re.FindStringSubmatch(evt.Content.AsMessage().FormattedBody)
// 			if len(match) == 1 {
// 				mtrx.quote(evt.RoomID, id.UserID(match[0]))
// 			} else {
// 				mtrx.sendMessage(evt.RoomID, fmt.Sprintf("Cannot find a recent message from %s", cmd[1]))
// 			}
// 		} else {
// 			mtrx.sendMessage(evt.RoomID, "Usage:\n!quote <user> - Quotes a users recent message\n!quote - Returns a random quote from this room")
// 			return
// 		}
// 	case "rss":
// 		output := mtrx.parseRSSCommand(cmd, evt.RoomID)
// 		mtrx.sendMessage(evt.RoomID, output)
// 	case "starthunt":
// 		fallthrough
// 	case "stophunt":
// 		fallthrough
// 	case "bang":
// 		mtrx.sendMessage(evt.RoomID, "Duckhunt coming soon...")
// 	case "ping":
// 		mtrx.sendMessage(evt.RoomID, "pong")
// 	}
// }

// func (mtrx *MtrxClient) sendMessage(roomID id.RoomID, text string) {
// 	if _, ok := mtrx.roomEncrypted[roomID]; !ok {
// 		_, err := mtrx.c.SendNotice(roomID, text)
// 		if err != nil {
// 			log.Print(err)
// 		}
// 		return
// 	}

// 	content := event.MessageEventContent{
// 		MsgType: "m.notice",
// 		Body:    text,
// 	}
// 	encrypted, err := mtrx.olm.EncryptMegolmEvent(roomID, event.EventMessage, content)
// 	// These three errors mean we have to make a new Megolm session
// 	if err == crypto.SessionExpired || err == crypto.SessionNotShared || err == crypto.NoGroupSession {
// 		err = mtrx.olm.ShareGroupSession(roomID, mtrx.getUserIDs(roomID))
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}
// 		encrypted, err = mtrx.olm.EncryptMegolmEvent(roomID, event.EventMessage, content)
// 	}
// 	if err != nil {
// 		log.Print(err)
// 		return
// 	}

// 	_, err = mtrx.c.SendMessageEvent(roomID, event.EventEncrypted, encrypted)
// 	if err != nil {
// 		log.Print(err)
// 		return
// 	}
// }

// func (mtrx *MtrxClient) getUserIDs(roomID id.RoomID) []id.UserID {
// 	members, err := mtrx.c.JoinedMembers(roomID)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	userIDs := make([]id.UserID, len(members.Joined))
// 	i := 0
// 	for userID := range members.Joined {
// 		userIDs[i] = userID
// 		i++
// 	}
// 	return userIDs
// }

// func (mtrx *MtrxClient) VerifySASMatch(otherDevice *crypto.DeviceIdentity, sas crypto.SASData) bool {
// 	return true
// }

// func (mtrx *MtrxClient) VerificationMethods() []crypto.VerificationMethod {
// 	return []crypto.VerificationMethod{
// 		crypto.VerificationMethodDecimal{},
// 	}
// }

// func (mtrx *MtrxClient) OnCancel(cancelledByUs bool, reason string, reasonCode event.VerificationCancelCode) {
// 	return
// }

// func (mtrx *MtrxClient) OnSuccess() {
// 	return
// }
