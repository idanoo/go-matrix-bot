package gomatrixbot

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var (
	// Matrix creds
	MatrixHost     string
	MatrixUsername string
	MatrixPassword string
	Admin          string
	Debug          bool

	BotDb     = "/config/bot.db"
	MautrixDB = "/config/mautrix.db"
)

type MtrxClient struct {
	c           *mautrix.Client
	startTime   int64
	db          *sql.DB
	quitMeDaddy chan struct{}
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

	// Boot recent messages
	go initRecentMessages()

	// Connect new client to matrix
	client, err := mautrix.NewClient(MatrixHost, "", "")
	if err != nil {
		panic(err)
	}

	mtrx.c = client
	mtrx.startTime = time.Now().UnixMilli()

	log := zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.Stamp
	})).With().Timestamp().Logger()
	if !Debug {
		log = log.Level(zerolog.InfoLevel)
	}
	mtrx.c.Log = log
	syncer := mtrx.c.Syncer.(*mautrix.DefaultSyncer)

	// On incoming message
	syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
		log.Info().
			Str("sender", evt.Sender.String()).
			Str("body", evt.Content.AsMessage().Body).
			Msg(evt.Type.String())

		// Thread for s p e e d
		go mtrx.handleMessageEvent(source, evt)
	})

	// On invite - Auto accept!
	syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		if evt.GetStateKey() == client.UserID.String() && evt.Content.AsMember().Membership == event.MembershipInvite {
			_, err := client.JoinRoomByID(evt.RoomID)
			if err == nil {
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

	cryptoHelper.LoginAs = &mautrix.ReqLogin{
		Type:       mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: MatrixUsername},
		Password:   MatrixPassword,
		DeviceID:   "GoMatrixBot",
	}

	err = cryptoHelper.Init()
	if err != nil {
		panic(err)
	}
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

	// forever loop until quit
	mtrx.quitMeDaddy = make(chan struct{})
	// run every damn minute
	everyDamnMinute := time.NewTimer(time.Minute)
	for {
		select {
		case <-mtrx.quitMeDaddy:
			log.Info().Msg("Received quit command!!!")
			cancelSync()
			syncStopWait.Wait()
			err = cryptoHelper.Close()
			if err != nil {
				log.Error().Err(err).Msg("Error closing database")
			}
			os.Exit(0)
			break

		case <-everyDamnMinute.C:
			go mtrx.purgeOldMessages(time.Now().Add(-5 * time.Minute))
		}
	}

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

func (mtrx *MtrxClient) handleMessageEvent(source mautrix.EventSource, evt *event.Event) {
	// If parsing own messages.. stop that too
	if evt.Sender.String() == mtrx.c.UserID.String() {
		return
	}

	// If syncing older messages.. stop that now
	if evt.Timestamp < mtrx.startTime {
		return
	}

	// Parse command if intended for us (. commands only!)
	if evt.Content.AsMessage().Body[0:1] == "." {
		mtrx.parseCommand(source, evt)
		return
	} else {
		// Add msg!
		addRecentMessage(evt.RoomID.String(), evt.ID.String(), evt.Sender.String(), evt.Content.AsMessage().Body)
	}
}

func (mtrx *MtrxClient) parseCommand(source mautrix.EventSource, evt *event.Event) {
	cmd := strings.Split(strings.TrimSpace(evt.Content.AsMessage().Body[1:]), " ")
	if len(cmd) == 0 {
		return
	}

	switch cmd[0] {
	case "ping":
		mtrx.sendMessage(evt.RoomID, "pong")
	case "echo":
		mtrx.sendMessage(evt.RoomID, strings.Join(cmd[1:], " "))
	case "quit":
		mtrx.sendMessage(evt.RoomID, "Shutting down... maybe?")
		// <-mtrx.quitMeDaddy
	case "s":
		msg := mtrx.sedThis(string(evt.RoomID), strings.Join(cmd[1:], " "))
		mtrx.sendMessage(evt.RoomID, msg)
	case "recent":
		if evt.Sender.Homeserver() != "deepspace.cafe" {
			mtrx.sendMessage(evt.RoomID, evt.Sender.Homeserver()+" not whitelisted")
			return
		}

		msgs := mtrx.getRecentMessages(evt.RoomID.String(), 10)
		mtrx.sendMessage(evt.RoomID, strings.Join(msgs, "\n\n"))
		// case "quote":
		// 	if len(cmd) == 1 {
		// 		user, quote := mtrx.getRandomQuote(evt.RoomID)
		// 		mtrx.sendMessage(evt.RoomID, fmt.Sprintf("Quote from %s: %s", user, quote))
		// 		return
		// 	} else if len(cmd) == 2 {
		// 		re := regexp.MustCompile(`@.*\.[^"]+`)
		// 		match := re.FindStringSubmatch(evt.Content.AsMessage().FormattedBody)
		// 		if len(match) == 1 {
		// 			mtrx.quote(evt.RoomID, id.UserID(match[0]))
		// 		} else {
		// 			mtrx.sendMessage(evt.RoomID, fmt.Sprintf("Cannot find a recent message from %s", cmd[1]))
		// 		}
		// 	} else {
		// 		mtrx.sendMessage(evt.RoomID, "Usage:\n!quote <user> - Quotes a users recent message\n!quote - Returns a random quote from this room")
		// 		return
		// 	}
		// case "rss":
		// 	output := mtrx.parseRSSCommand(cmd, evt.RoomID)
		// 	mtrx.sendMessage(evt.RoomID, output)
		// case "starthunt":
		// 	fallthrough
		// case "stophunt":
		// 	fallthrough
		// case "bang":
		// 	mtrx.sendMessage(evt.RoomID, "Duckhunt coming soon...")

	}
}

func (mtrx *MtrxClient) sendMessage(roomID id.RoomID, text string) {
	resp, err := mtrx.c.SendText(roomID, text)
	if err != nil {
		mtrx.c.Log.Error().Err(err).Msg("Failed to send event")
	} else {
		mtrx.c.Log.Info().Str("event_id", resp.EventID.String()).Msg("Event sent")
	}
}

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
