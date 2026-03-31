package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Database wraps database connection and queries
type Database struct {
	*sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(dsn string) (*Database, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &Database{db}, nil
}
