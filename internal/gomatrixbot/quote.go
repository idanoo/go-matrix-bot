package gomatrixbot

import (
	"fmt"
	"log"
	"strings"

	"maunium.net/go/mautrix/id"
)

// This is per roomID
type quoteCache struct {
	lastMessage map[string]string
}

// Create empty map
func (mtrx *MtrxClient) initQuote() {
	mtrx.quotes = make(map[id.RoomID]quoteCache)
}

func (mtrx *MtrxClient) appendLastMessage(roomID id.RoomID, userID id.UserID, message string) {
	userIDLocal := getLocalUserPart(userID.String())
	if _, roomOk := mtrx.quotes[roomID]; !roomOk {
		last := make(map[string]string)
		last[userIDLocal] = message
		mtrx.quotes[roomID] = quoteCache{
			lastMessage: last,
		}
	} else {
		mtrx.quotes[roomID].lastMessage[userIDLocal] = message
	}
}

// Grab last message from user and store it in DB
func (mtrx *MtrxClient) quote(roomID id.RoomID, userID string) {
	if roomVal, roomOk := mtrx.quotes[roomID]; roomOk {
		if userVal, userOk := roomVal.lastMessage[userID]; userOk {
			// Store to quote DB
			mtrx.storeQuote(roomID, userID, userVal)

			// Send quote
			_, err := mtrx.c.SendNotice(roomID, fmt.Sprintf("Quoted %s: %s", userID, userVal))
			if err != nil {
				log.Print(err)
			}
			return
		}
	}

	_, err := mtrx.c.SendNotice(roomID, fmt.Sprintf("Cannot find a recent message from %s", userID))
	if err != nil {
		log.Print(err)
	}
}

func getLocalUserPart(str string) string {
	parts := strings.Split(str, ":")

	return parts[0][1:]
}
