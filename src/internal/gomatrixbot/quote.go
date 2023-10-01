package gomatrixbot

// import (
// 	"fmt"
// 	"strings"

// 	"maunium.net/go/mautrix/id"
// )

// // This is per roomID
// type quoteCache struct {
// 	lastMessage map[id.UserID]string
// }

// // Create empty map
// func (mtrx *MtrxClient) initQuote() {
// 	mtrx.quotes = make(map[id.RoomID]quoteCache)
// }

// func (mtrx *MtrxClient) appendLastMessage(roomID id.RoomID, userID id.UserID, message string) {
// 	if _, roomOk := mtrx.quotes[roomID]; !roomOk {
// 		last := make(map[id.UserID]string)
// 		last[userID] = message
// 		mtrx.quotes[roomID] = quoteCache{
// 			lastMessage: last,
// 		}
// 	} else {
// 		mtrx.quotes[roomID].lastMessage[userID] = message
// 	}
// }

// // Grab last message from user and store it in DB
// func (mtrx *MtrxClient) quote(roomID id.RoomID, userID id.UserID) error {
// 	if roomVal, roomOk := mtrx.quotes[roomID]; roomOk {
// 		if userVal, userOk := roomVal.lastMessage[userID]; userOk {
// 			// Store to quote DB
// 			mtrx.storeQuote(roomID, userID, userVal)

// 			// Send quote
// 			mtrx.sendMessage(roomID, fmt.Sprintf("Quoted %s: %s", userID, userVal))
// 			return nil
// 		}
// 	}

// 	mtrx.sendMessage(roomID, fmt.Sprintf("Cannot find a recent message from %s", userID))

// 	return nil
// }

// func getLocalUserPart(str string) string {
// 	parts := strings.Split(str, ":")

// 	return parts[0][1:]
// }
