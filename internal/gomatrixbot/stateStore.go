package gomatrixbot

import (
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// Simple crypto.StateStore implementation that says all rooms are encrypted.
type stateStore struct{}

var _ crypto.StateStore = &stateStore{}

func (fss *stateStore) IsEncrypted(roomID id.RoomID) bool {
	return true
}

func (fss *stateStore) GetEncryptionEvent(roomID id.RoomID) *event.EncryptionEventContent {
	return &event.EncryptionEventContent{
		Algorithm:              id.AlgorithmMegolmV1,
		RotationPeriodMillis:   7 * 24 * 60 * 60 * 1000,
		RotationPeriodMessages: 100,
	}
}

func (fss *stateStore) FindSharedRooms(userID id.UserID) []id.RoomID {
	return []id.RoomID{}
}
