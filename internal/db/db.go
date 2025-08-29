package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool // unexported

// ConnectDB initializes the Postgres connection pool
func ConnectDB(connStr string) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatal("Failed to parse connection string: ", err)
	}

	// Optional: tune pool settings
	config.MaxConns = 100
	config.MinConns = 2
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	p, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal("Unable to connect to database: ", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := p.Ping(ctx); err != nil {
		log.Fatal("Unable to ping database: ", err)
	}

	pool = p
	log.Println("Database connected successfully!")
}

// GetDB returns the pool for external usage
func GetDB() *pgxpool.Pool {
	return pool
}

// CloseDB closes the pool when the app shuts down
func CloseDB() {
	if pool != nil {
		pool.Close()
		log.Println("Database connection pool closed.")
	}
}
