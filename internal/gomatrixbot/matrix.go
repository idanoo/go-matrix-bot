package gomatrixbot

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
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
	c           *mautrix.Client
	startTime   int64
	db          *sql.DB
	olm         *crypto.OlmMachine
	cryptoStore *crypto.SQLCryptoStore

	ducks         duckHunt
	quotes        map[id.RoomID]quoteCache
	roomEncrypted map[id.RoomID]bool
}

// Run - starts bot!
func Run() {
	mtrx := MtrxClient{}

	// Get DB connection and run any pending migrations
	mtrx.db = InitDb()
	defer mtrx.CloseDbConn()

	// Connet to matrix!
	mtrx.initClient()
	mtrx.startTime = time.Now().UnixMilli()

	mtrx.roomEncrypted = make(map[id.RoomID]bool)

	syncer := mtrx.c.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnSync(func(resp *mautrix.RespSync, since string) bool {
		mtrx.olm.ProcessSyncResponse(resp, since)
		return true
	})

	syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		mtrx.olm.HandleMemberEvent(evt)
	})

	syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		go mtrx.handleInvite(source, evt)
	})

	syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
		mtrx.roomEncrypted[evt.RoomID] = false
		go mtrx.handleEvent(source, evt)
	})

	syncer.OnEventType(event.EventEncrypted, func(source mautrix.EventSource, evt *event.Event) {
		mtrx.roomEncrypted[evt.RoomID] = true
		if evt.Timestamp < mtrx.startTime {
			return
		}

		decrypted, err := mtrx.olm.DecryptMegolmEvent(evt)
		if err != nil {
			fmt.Println(err)
		} else {
			go mtrx.handleEvent(source, decrypted)
		}

	})

	log.Printf("Started gomatrixbot")

	// Init all the things
	// go mtrx.initDuckHunt()
	go mtrx.initQuote()
	go mtrx.initRss()

	// Launch'er up
	err := mtrx.c.Sync()
	if err != nil {
		log.Fatal(err)
	}
}

func (mtrx *MtrxClient) initClient() {
	// Connect new client
	client, err := mautrix.NewClient(MatrixHost, "", "")
	if err != nil {
		log.Fatal(err)
	}

	// Login
	resp, err := client.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: MatrixUsername},
		DeviceID:         "dbot",
		Password:         MatrixPassword,
		StoreCredentials: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	accountID := resp.UserID.String() + "-" + resp.DeviceID.String()
	cryptoStore := crypto.NewSQLCryptoStore(mtrx.db, "sqlite3", accountID, resp.DeviceID, []byte(client.DeviceID.String()+"pickle"), &logger{})
	if err != nil {
		log.Fatal(err)
	}

	if err = cryptoStore.CreateTables(); err != nil {
		log.Println(err)
	}

	mtrx.cryptoStore = cryptoStore

	mach := crypto.NewOlmMachine(client, &logger{}, mtrx.cryptoStore, &stateStore{})
	mach.AcceptVerificationFrom = func(_ string, otherDevice *crypto.DeviceIdentity, _ id.RoomID) (crypto.VerificationRequestResponse, crypto.VerificationHooks) {
		return crypto.AcceptRequest, mtrx
	}

	err = mach.Load()
	if err != nil {
		log.Fatal(err)
	}

	mtrx.c = client
	mtrx.olm = mach
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

	if evt.Sender == mtrx.c.UserID {
		return
	}

	content := struct {
		Inviter id.UserID `json:"inviter"`
	}{evt.Sender}

	fmt.Println("Invite by" + evt.Sender.String())

	if _, err := mtrx.c.JoinRoom(evt.RoomID.String(), "", content); err != nil {
		fmt.Printf("Failed to join room: %s", evt.RoomID.String())
	} else {
		fmt.Printf("Joined room: %s", evt.RoomID.String())
	}
}

func (mtrx *MtrxClient) parseCommand(source mautrix.EventSource, evt *event.Event) {
	cmd := strings.Split(strings.TrimSpace(evt.Content.AsMessage().Body[1:]), " ")
	if len(cmd) == 0 {
		return
	}

	switch cmd[0] {
	case "help":
		msg := "@dbot:mtrx.nz commands:\n" +
			"\n!echo - echo message back to channel" +
			"\n!quote <user>, !quote - quote users last message or returns a random quote" +
			"\n!starthunt, !stophunt, !bang - duckhunt commands" +
			"\n!rss - RSS subscription commands"
		mtrx.sendMessage(evt.RoomID, msg)
	case "echo":
		mtrx.sendMessage(evt.RoomID, strings.Join(cmd[1:], " "))
	case "quote":
		if len(cmd) == 1 {
			user, quote := mtrx.getRandomQuote(evt.RoomID)
			mtrx.sendMessage(evt.RoomID, fmt.Sprintf("Quote from %s: %s", user, quote))
			return
		} else if len(cmd) == 2 {
			re := regexp.MustCompile(`@.*\.[^"]+`)
			match := re.FindStringSubmatch(evt.Content.AsMessage().FormattedBody)
			if len(match) == 1 {
				mtrx.quote(evt.RoomID, id.UserID(match[0]))
			} else {
				mtrx.sendMessage(evt.RoomID, fmt.Sprintf("Cannot find a recent message from %s", cmd[1]))
			}
		} else {
			mtrx.sendMessage(evt.RoomID, "Usage:\n!quote <user> - Quotes a users recent message\n!quote - Returns a random quote from this room")
			return
		}
	case "rss":
		output := mtrx.parseRSSCommand(cmd, evt.RoomID)
		mtrx.sendMessage(evt.RoomID, output)
	case "starthunt":
		fallthrough
	case "stophunt":
		fallthrough
	case "bang":
		mtrx.sendMessage(evt.RoomID, "Duckhunt coming soon...")
	case "ping":
		mtrx.sendMessage(evt.RoomID, "pong")
	}
}

func (mtrx *MtrxClient) sendMessage(roomID id.RoomID, text string) {
	if _, ok := mtrx.roomEncrypted[roomID]; !ok {
		_, err := mtrx.c.SendNotice(roomID, text)
		if err != nil {
			log.Print(err)
		}
		return
	}

	content := event.MessageEventContent{
		MsgType: "m.notice",
		Body:    text,
	}
	encrypted, err := mtrx.olm.EncryptMegolmEvent(roomID, event.EventMessage, content)
	// These three errors mean we have to make a new Megolm session
	if err == crypto.SessionExpired || err == crypto.SessionNotShared || err == crypto.NoGroupSession {
		err = mtrx.olm.ShareGroupSession(roomID, mtrx.getUserIDs(roomID))
		if err != nil {
			log.Println(err)
			return
		}
		encrypted, err = mtrx.olm.EncryptMegolmEvent(roomID, event.EventMessage, content)
	}
	if err != nil {
		log.Print(err)
		return
	}

	_, err = mtrx.c.SendMessageEvent(roomID, event.EventEncrypted, encrypted)
	if err != nil {
		log.Print(err)
		return
	}
}

func (mtrx *MtrxClient) getUserIDs(roomID id.RoomID) []id.UserID {
	members, err := mtrx.c.JoinedMembers(roomID)
	if err != nil {
		log.Fatal(err)
	}

	userIDs := make([]id.UserID, len(members.Joined))
	i := 0
	for userID := range members.Joined {
		userIDs[i] = userID
		i++
	}
	return userIDs
}

func (mtrx *MtrxClient) VerifySASMatch(otherDevice *crypto.DeviceIdentity, sas crypto.SASData) bool {
	return true
}

func (mtrx *MtrxClient) VerificationMethods() []crypto.VerificationMethod {
	return []crypto.VerificationMethod{
		crypto.VerificationMethodDecimal{},
	}
}

func (mtrx *MtrxClient) OnCancel(cancelledByUs bool, reason string, reasonCode event.VerificationCancelCode) {
	return
}

func (mtrx *MtrxClient) OnSuccess() {
	return
}
