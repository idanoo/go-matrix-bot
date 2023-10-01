package gomatrixbot

// import (
// 	"fmt"
// 	"log"
// 	"time"

// 	"github.com/mmcdole/gofeed"
// 	"maunium.net/go/mautrix/id"
// )

// type RSSFeed struct {
// 	URL         string
// 	RoomID      string
// 	LastUpdated int64
// }

// var (
// 	rssFeedUpdateTime = int64(300)
// )

// // Starts a poller to get updates!
// func (mtrx *MtrxClient) initRss() {
// 	everyMinute := time.NewTimer(time.Minute)
// 	func() {
// 		<-everyMinute.C
// 		go mtrx.pollFeeds()
// 	}()
// }

// func (mtrx *MtrxClient) parseRSSCommand(cmd []string, roomID id.RoomID) string {
// 	defaultText := "Usage:\n!rss add <url> - Adds RSS sub" +
// 		"\n!rss list - Lists feeds" +
// 		"\n!rss remove <url>"

// 	if len(cmd) == 2 && cmd[1] == "list" {
// 		return mtrx.listRSSFeeds(roomID)
// 	} else if len(cmd) == 3 {
// 		if cmd[1] == "add" {
// 			return mtrx.addRSSFeed(cmd[2], roomID)
// 		} else if cmd[1] == "remove" {
// 			return mtrx.removeRSSFeed(cmd[2], roomID)
// 		} else {
// 			return defaultText
// 		}
// 	} else {
// 		return defaultText
// 	}
// }

// // Add new feed
// func (mtrx *MtrxClient) addRSSFeed(feedURL string, roomID id.RoomID) string {
// 	err := mtrx.addDBRSSFeed(feedURL, roomID)
// 	if err != nil {
// 		log.Println(err)
// 		return "Error adding RSS feed. Feed already exists?"
// 	}

// 	return "Added RSS feed " + feedURL
// }

// // List all feeds for room
// func (mtrx *MtrxClient) removeRSSFeed(feedURL string, roomID id.RoomID) string {
// 	err := mtrx.removeDBRSSFeed(feedURL, roomID)
// 	if err != nil {
// 		log.Println(err)
// 		return "Error removing RSS feed."
// 	}

// 	return "Removed RSS feed " + feedURL
// }

// // List all feeds for room
// func (mtrx *MtrxClient) listRSSFeeds(roomID id.RoomID) string {
// 	feeds, err := mtrx.listDBRSSFeed(roomID)
// 	if err != nil {
// 		log.Println(err)
// 		return "Error listing RSS feeds."
// 	}

// 	if len(feeds) == 0 {
// 		return "No RSS feeds added"
// 	}

// 	output := "Current RSS feeds:\n"
// 	for _, feed := range feeds {
// 		output = output + "\n" + feed
// 	}
// 	return output
// }

// // Poll all feeds
// func (mtrx *MtrxClient) pollFeeds() {
// 	log.Println("polling rss")
// 	// Get all feeds with due update time

// 	feeds := mtrx.listAllDBRSSFeeds()
// 	for _, feed := range feeds {
// 		// If polled since our time, skip it (now - 10min)
// 		if feed.LastUpdated > time.Now().Unix()-rssFeedUpdateTime {
// 			continue
// 		}

// 		fp := gofeed.NewParser()
// 		rss, err := fp.ParseURL(feed.URL)
// 		if err != nil {
// 			mtrx.sendMessage(id.RoomID(feed.RoomID), "Failed to parse RSS feed. Removing "+feed.URL)
// 			_ = mtrx.removeRSSFeed(feed.URL, id.RoomID(feed.RoomID))
// 			continue
// 		}

// 		for _, item := range rss.Items {
// 			if item.PublishedParsed == nil {
// 				log.Printf("Skipping %s. Nil published date", feed.URL)
// 				continue
// 			}

// 			// Don't post if we've polled past it
// 			if item.PublishedParsed.Unix() <= feed.LastUpdated {
// 				log.Println("Skipping RSS item: " + item.Title)
// 				continue
// 			}

// 			// Don't post older than an hour!
// 			if item.PublishedParsed.Unix() <= time.Now().Unix()-3600 {
// 				continue
// 			}

// 			mtrx.sendMessage(id.RoomID(feed.RoomID), fmt.Sprintf("%s - %s\n%s", rss.Title, item.Title, item.Link))
// 		}

// 		go mtrx.updateDBRSSFeed(feed.RoomID, feed.URL)
// 	}
// 	// Post each update
// }
