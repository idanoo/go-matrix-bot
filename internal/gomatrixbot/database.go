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

	db *sql.DB
)

func InitDb() {
	dbConn, err := sql.Open("sqlite3", "file:"+DBFile+"?loc=auto")
	if err != nil {
		panic(err)
	}

	dbConn.SetConnMaxLifetime(0)
	dbConn.SetMaxOpenConns(20)
	dbConn.SetMaxIdleConns(2)

	err = dbConn.Ping()
	if err != nil {
		panic(err)
	}

	db = dbConn

	runMigrations()
}

// CloseDbConn - Closes DB connection
func CloseDbConn() {
	db.Close()
}

func runMigrations() {
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
