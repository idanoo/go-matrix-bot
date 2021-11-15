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

var (
	DBFile string
)

func InitDb() *sql.DB {
	db, err := sql.Open("sqlite3", "file:"+DBFile+"?loc=auto")
	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxOpenConns(20)
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
func (mtrx *MtrxClient) storeQuote(roomID id.RoomID, userID string, quote string) {
	_, err := mtrx.db.Exec(
		"INSERT INTO `quotes` (`user_id`, `room_id`, `timestamp`, `quote`) VALUES (?,?,?,?)",
		userID, roomID.String(), time.Now().Unix(), quote)
	if err != nil {
		log.Print(err)
	}
}
