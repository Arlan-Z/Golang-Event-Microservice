package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var direction, dbPath, migrationsPath string

	flag.StringVar(&direction, "direction", "up", "Migration direction (up or down)")
	flag.StringVar(&dbPath, "dbpath", "./events.db", "Path to the SQLite database file")
	flag.StringVar(&migrationsPath, "path", "migrations", "Path to migrations folder")
	flag.Parse()

	if direction != "up" && direction != "down" {
		log.Fatal("Direction must be 'up' or 'down'")
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			log.Fatalf("Failed to create database file: %v", err)
		}
		file.Close()
		log.Printf("Created database file: %s", dbPath)
	} else if err != nil {
		log.Fatalf("Error checking database file: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			log.Printf("Warning: failed to close database connection: %v", cerr)
		}
	}()
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection opened successfully.")

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatalf("Could not create migrate driver instance: %v", err)
	}
	log.Println("Migrate database driver instance created.")

	sourceURL := fmt.Sprintf("file://%s", migrationsPath)

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"sqlite3",
		driver,
	)
	if err != nil {
		log.Fatalf("Migration instance creation failed: %v", err)
	}
	log.Println("Migrate instance created successfully.")

	var migrateErr error
	if direction == "up" {
		log.Println("Applying migrations up...")
		migrateErr = m.Up()
	} else {
		log.Println("Applying migrations down...")
		migrateErr = m.Down()
	}

	if migrateErr != nil && !errors.Is(migrateErr, migrate.ErrNoChange) {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			log.Printf("Error closing migration source: %v", sourceErr)
		}
		if dbErr != nil {
			log.Printf("Error closing migration database connection handle: %v", dbErr)
		}
		log.Fatalf("Migration failed: %v", migrateErr)
	}

	if errors.Is(migrateErr, migrate.ErrNoChange) {
		log.Println("No migrations to apply.")
	} else {
		log.Println("Migrations applied successfully.")
	}

	sourceErr, dbErr := m.Close()
	if sourceErr != nil {
		log.Printf("Error closing migration source: %v", sourceErr)
	}
	if dbErr != nil {
		log.Printf("Error closing migration database connection handle: %v", dbErr)
	}
	log.Println("Migration resources closed.")
}
