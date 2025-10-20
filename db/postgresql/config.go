package postgresql

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// ConnectDB establishes a connection to the database using the DATABASE_URL from .env
func ConnectDB() (*sqlx.DB, error) { //function open connect

	// fetch the DATABASE_URL from the environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required but not set")
	}

	// connect to the PostgreSQL database using sqlx
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	// set connection pool settings
	db.SetMaxIdleConns(10)   // set maximum idle connections
	db.SetMaxOpenConns(10)   // set maximum open connections (including idle)
	db.SetConnMaxLifetime(0) // unlimited connection lifetime (0 means no limit)

	fmt.Println("Database connected!")
	return db, nil
}
