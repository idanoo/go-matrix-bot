package gomatrixbot

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
	"maunium.net/go/mautrix/id"
)

func InitDb() *sql.DB {
	db, err := sql.Open("sqlite3", "file:"+BotDb+"?loc=auto")
	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	runMigrations(db)

	return db
}

// CloseDbConn - Closes DB connection
func (mtrx *MtrxClient) CloseDbConn() {
	mtrx.db.Close()
}

func runMigrations(db *sql.DB) {
	fmt.Println("Checking database migrations")
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		panic(fmt.Sprintf("Unable to run DB Migrations %v", err))
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		panic(fmt.Sprintf("Error fetching DB Migrations %v", err))
	}

	err = m.Up()
	if err != nil {
		// Skip 'no change'. This is fine. Everything is fine.
		if err.Error() == "no change" {
			fmt.Println("Database already up to date")
			return
		}

		panic(fmt.Sprintf("Error running DB Migrations %v", err))
	}

	fmt.Println("Database migrations complete")
}

func (mtrx *MtrxClient) getDuckHunt() map[string]int64 {
	results := make(map[string]int64)
	row, err := mtrx.db.Query("SELECT `room_id`, `enabled` FROM `duck_hunt`")
	if err != sql.ErrNoRows {
		return results
	} else if err != nil {
		log.Fatal(err)
	}

	defer row.Close()
	for row.Next() {
		var room_id string
		var enabled int64
		row.Scan(&room_id, &enabled)
		results[room_id] = enabled
	}

	return results
}

// QUOTE
func (mtrx *MtrxClient) getRandomQuote(roomID id.RoomID) (string, string) {
	userID := "N/A"
	quote := "No quotes for this room"
	row, err := mtrx.db.Query("SELECT `quote`, `user_id` FROM `quotes` WHERE `room_id` = ? ORDER BY RANDOM() LIMIT 1", roomID)
	if err == sql.ErrNoRows {
		return userID, quote
	} else if err != nil {
		log.Print(err)
		return userID, quote
	}

	defer row.Close()
	for row.Next() {
		row.Scan(&quote, &userID)
	}

	return userID, quote
}

// QUOTE
func (mtrx *MtrxClient) storeQuote(roomID id.RoomID, userID id.UserID, quote string) {
	_, err := mtrx.db.Exec(
		"INSERT INTO `quotes` (`user_id`, `room_id`, `timestamp`, `quote`) VALUES (?,?,?,?)",
		userID.String(), roomID.String(), time.Now().Unix(), quote)
	if err != nil {
		log.Print(err)
	}
}

// Add feed to DB
func (mtrx *MtrxClient) addDBRSSFeed(url string, roomID id.RoomID) error {
	_, err := mtrx.db.Exec(
		"INSERT INTO `rss_feeds` (`room_id`, `url`, `last_updated`) VALUES (?,?,?)",
		roomID.String(), url, time.Now().Unix()-600)

	return err
}

// Remove feed from DB
func (mtrx *MtrxClient) removeDBRSSFeed(url string, roomID id.RoomID) error {
	_, err := mtrx.db.Exec(
		"DELETE FROM `rss_feeds` WHERE `room_id` = ? AND `url` = ?",
		roomID.String(), url)

	return err
}

// Get feeds for room
func (mtrx *MtrxClient) listDBRSSFeed(roomID id.RoomID) ([]string, error) {
	feeds := []string{}

	row, err := mtrx.db.Query("SELECT `url` FROM `rss_feeds` WHERE `room_id` = ?", roomID.String())
	if err == sql.ErrNoRows {
		return feeds, nil
	} else if err != nil {
		return feeds, err
	}

	defer row.Close()
	for row.Next() {
		var url string
		row.Scan(&url)
		feeds = append(feeds, url)
	}

	return feeds, nil
}

// Get all feeds
// func (mtrx *MtrxClient) listAllDBRSSFeeds() []RSSFeed {
// 	feeds := []RSSFeed{}

// 	row, err := mtrx.db.Query("SELECT `url`, `room_id`, `last_updated` FROM `rss_feeds`")
// 	if err == sql.ErrNoRows {
// 		return feeds
// 	} else if err != nil {
// 		log.Println(err)
// 		return feeds
// 	}

// 	defer row.Close()
// 	for row.Next() {
// 		var rssFeed RSSFeed
// 		row.Scan(&rssFeed.URL, &rssFeed.RoomID, &rssFeed.LastUpdated)
// 		feeds = append(feeds, rssFeed)
// 	}

// 	return feeds
// }

// func (mtrx *MtrxClient) updateDBRSSFeed(roomID string, url string) {
// 	_, err := mtrx.db.Exec(
// 		"UPDATE `rss_feeds` SET `last_updated` = UNIX_TIMESTAMP() WHERE `room_id` = ? AND `url` = ?)",
// 		roomID, url)

// 	if err != nil {
// 		fmt.Println(err)
// 	}
// }
