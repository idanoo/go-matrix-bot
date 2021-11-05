package gomatrixbot

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
)

var (
	DBFile string

	db DB
)

type DB struct {
	dbconn *sql.DB
}

func InitDb() {
	conn, err := sql.Open("sqlite3", "file:"+DBFile+"?loc=auto")
	if err != nil {
		panic(err)
	}

	conn.SetConnMaxLifetime(0)
	conn.SetMaxOpenConns(20)
	conn.SetMaxIdleConns(2)

	err = conn.Ping()
	if err != nil {
		panic(err)
	}

	db = DB{
		dbconn: conn,
	}

	runMigrations()
}

// CloseDbConn - Closes DB connection
func CloseDbConn() {
	db.dbconn.Close()
}

func runMigrations() {
	fmt.Println("Checking database migrations")
	driver, err := sqlite3.WithInstance(db.dbconn, &sqlite3.Config{})
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

func (db *DB) getRooms() map[string]string {
	results := make(map[string]string)
	row, err := db.dbconn.Query("SELECT `id`, `name` FROM `rooms`")
	if err != nil {
		log.Fatal(err)
	}

	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var id string
		var name string
		row.Scan(&id, &name)
		results[id] = name
	}

	return results
}
