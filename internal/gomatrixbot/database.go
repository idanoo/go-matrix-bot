package gomatrixbot

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
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

// func (db *DB) getRooms() map[string]string {
// 	results := make(map[string]string)
// 	row, err := db.dbconn.Query("SELECT `id`, `name` FROM `rooms`")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	defer row.Close()
// 	for row.Next() { // Iterate and fetch the records from result cursor
// 		var id string
// 		var name string
// 		row.Scan(&id, &name)
// 		results[id] = name
// 	}

// 	return results
// }
