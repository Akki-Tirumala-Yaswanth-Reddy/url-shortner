package db

import (
    "context"
    "fmt"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/joho/godotenv"
)

var DB *pgxpool.Pool

func InitDB() error {
    

    dbURL := os.Getenv("DATABASE_URL")

    if dbURL == "" {
        if err := godotenv.Load(".env"); err != nil {
            return fmt.Errorf("failed to load .env: %w", err)
        }
        dbURL = os.Getenv("DATABASE_URL")
    }

    if dbURL == "" {
        return fmt.Errorf("DATABASE_URL is empty")
    }

    pool, err := pgxpool.New(context.Background(), dbURL)
    if err != nil {
        return fmt.Errorf("unable to connect to database: %w", err)
    }

    if err := pool.Ping(context.Background()); err != nil {
        pool.Close()
        return fmt.Errorf("db ping failed: %w", err)
    }

    DB = pool
    return nil
}