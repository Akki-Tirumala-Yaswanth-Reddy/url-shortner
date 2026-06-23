package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var DB *pgxpool.Pool
var RDB *redis.Client

func loadEnv() {
	if os.Getenv("DATABASE_URL") == "" {
		godotenv.Load(".env")
	}
}

func InitDB() error {
	loadEnv()
	dbURL := os.Getenv("DATABASE_URL")

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

func InitRedis() error {
	loadEnv()
	redisURL := os.Getenv("REDIS_URL")

	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("invalid REDIS_URL: %w", err)
	}

	client := redis.NewClient(opt)
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	RDB = client
	return nil
}
