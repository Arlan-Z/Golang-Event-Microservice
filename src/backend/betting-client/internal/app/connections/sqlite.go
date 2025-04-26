package connections

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // Import driver
)

func NewSQLiteConnection(dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite database: %w", err)
	}

	// Optional: Configure connection pool settings if needed
	// db.SetMaxOpenConns(10)
	// db.SetMaxIdleConns(5)
	// db.SetConnMaxLifetime(time.Hour)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	log.Println("Successfully connected to SQLite database.")
	return db, nil
}
