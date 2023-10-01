package gomatrixbot

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// map[channel]map[timestamp]
var msgCache *MessageCache

type MessageCache struct {
	channels map[string]RecentMessages
}

type RecentMessages struct {
	messages map[int64]Message
}

type Message struct {
	ID      string
	User    string
	Content string
}

// Keep cache
func initRecentMessages() {
	msgs := make(map[string]RecentMessages)
	recentMessages := MessageCache{
		channels: msgs,
	}
	msgCache = &recentMessages
}

// add message to cache
func addRecentMessage(channelID string, messageID string, messageUser string, message string) {
	// Create chan if not exists
	if _, ok := msgCache.channels[channelID]; !ok {
		msgs := make(map[int64]Message)
		temp := RecentMessages{
			messages: msgs,
		}
		msgCache.channels[channelID] = temp
	}

	tempMessage := Message{
		ID:      messageID,
		User:    messageUser,
		Content: message,
	}

	msgCache.channels[channelID].messages[time.Now().UnixMilli()] = tempMessage

}

func (mtrx *MtrxClient) searchRecentMessages(channelID string, search string) (Message, error) {
	var msg Message
	if _, ok := msgCache.channels[channelID]; !ok {
		return msg, errors.New("No data for channel")
	} else if len(msgCache.channels[channelID].messages) == 1 {
		return msg, errors.New("No data for channel")
	}

	// fuck sorting
	sorted := msgCache.channels[channelID].messages
	keys := make([]int, 0, len(sorted))
	for k := range sorted {
		keys = append(keys, int(k))
	}

	sort.Sort(sort.Reverse(sort.IntSlice(keys[1:]))) // 1: to ignore the command we JUST issued

	for _, k := range keys {
		if strings.Contains(strings.ToLower(sorted[int64(k)].Content), strings.ToLower(search)) {
			return sorted[int64(k)], nil
		}
	}

	return msg, errors.New("Nothing found")
}

func (mtrx *MtrxClient) getRecentMessages(channelID string, num int64) []string {
	var vals []string
	i := 0
	if _, ok := msgCache.channels[channelID]; !ok {
		return append(vals, "No data for this channel")
	} else if len(msgCache.channels[channelID].messages) == 1 {
		return append(vals, "No data for this channel")
	}

	// fuck sorting
	sorted := msgCache.channels[channelID].messages
	keys := make([]int, 0, len(sorted))
	for k := range sorted {
		keys = append(keys, int(k))
	}

	sort.Sort(sort.Reverse(sort.IntSlice(keys[1:]))) // 1: to ignore the command we JUST issued

	for _, k := range keys {
		str := fmt.Sprintf("%s %s %s", time.UnixMilli(int64(k)).Format("02 Jan 06 15:04 -0700"), sorted[int64(k)].User, sorted[int64(k)].Content)
		vals = append(vals, str)
		i++
		if i > 5 {
			break
		}

	}

	return vals
}

// Purge before timestamp!
func (mtrx *MtrxClient) purgeOldMessages(before time.Time) {
	i := 0
	for channelID, data := range msgCache.channels {
		for timestamp, _ := range data.messages {
			ts := time.UnixMilli(timestamp)
			if ts.Before(before) {
				delete(msgCache.channels[channelID].messages, timestamp)
				i++
			}
		}
	}

	if i > 0 {
		mtrx.c.Log.Info().
			Str("msg", fmt.Sprintf("Cleaned up %d messages from cache", i)).
			Msg("EventCleanup")
	}
}
