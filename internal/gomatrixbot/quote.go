package gomatrixbot

import (
	"fmt"
	"log"

	"maunium.net/go/mautrix/id"
)

// This is per roomID
type quoteCache struct {
	lastMessage map[id.UserID]string
}

// Create empty map
func (mtrx *MtrxClient) initQuote() {
	mtrx.quotes = make(map[id.RoomID]quoteCache)
}

func (mtrx *MtrxClient) appendLastMessage(roomID id.RoomID, userID id.UserID, message string) {
	if _, roomOk := mtrx.quotes[roomID]; !roomOk {
		last := make(map[id.UserID]string)
		last[userID] = message
		mtrx.quotes[roomID] = quoteCache{
			lastMessage: last,
		}
	} else {
		mtrx.quotes[roomID].lastMessage[userID] = message
	}
}

// Grab last message from user and store it in DB
func (mtrx *MtrxClient) quote(roomID id.RoomID, userID id.UserID) {
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
