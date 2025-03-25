package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Initialize() error {
	ctx := context.Background()
	
	connStr := os.Getenv("DATABASE_URL")
	
	var err error
	DB, err = pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	
	if err := DB.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping the database: %w", err)
	}
	
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}