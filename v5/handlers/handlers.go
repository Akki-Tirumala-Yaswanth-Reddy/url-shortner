package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"
	"url_shortner/v5/db"
	"url_shortner/v5/helpers"
	"url_shortner/v5/models"
	"url_shortner/v5/analytics"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
)

func CreateShortUrl(w http.ResponseWriter, r *http.Request) {
	var req models.RequestJSONmodel
	var res models.ResponseJSONmodel

	if err := helpers.JSONdecoder(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.User == "" {
		http.Error(w, "user can't be empty", http.StatusBadRequest)
		return
	} else if req.Url == "" {
		http.Error(w, "url can't be empty", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	tx, err := db.DB.Begin(ctx)
	if err != nil {
		log.Println("begin tx:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var id int64
	insertQuery := `
        INSERT INTO urls (user_name, original_url)
        VALUES ($1, $2)
        RETURNING id;
    `
	if err := tx.QueryRow(ctx, insertQuery, req.User, req.Url).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		log.Println("insert:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortCode := helpers.EncodeBase62(uint64(id))

	updateQuery := `UPDATE urls SET short_code = $1 WHERE id = $2;`
	if _, err := tx.Exec(ctx, updateQuery, shortCode, id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "short code collision", http.StatusConflict)
			return
		}
		log.Println("update short_code:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("commit:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := db.RDB.Set(ctx, "url:"+shortCode, req.Url, 24*time.Hour).Err(); err != nil {
		log.Println("redis set:", err)
	}

	res.Id = id
	res.Url = "localhost:8080/redirect/" + shortCode

	if err := helpers.JSONencoder(w, &res, http.StatusOK); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Redirect(ana *analytics.AnalyticsObj) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("short_code")
		if shortCode == "" {
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		cacheKey := "url:" + shortCode

		// Check Redis cache first
		originalURL, err := db.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			log.Println("cache hit for:", shortCode)
			ana.Add(shortCode)
			http.Redirect(w, r, originalURL, http.StatusFound)
			return
		}
		if err != redis.Nil {
			log.Println("redis get:", err)
		}

		// Cache miss — query database
		var url models.Url
		query := `
			SELECT user_name, short_code, original_url
			FROM urls
			WHERE short_code=$1;
		`
		if err := db.DB.QueryRow(ctx, query, shortCode).Scan(&url.Username, &url.Short_code, &url.Original_url); err != nil {
			log.Println(err.Error())
			http.Error(w, "url not found", http.StatusNotFound)
			return
		}

		// Populate cache for subsequent requests
		if err := db.RDB.Set(ctx, cacheKey, url.Original_url, 24*time.Hour).Err(); err != nil {
			log.Println("redis set:", err)
		}
		ana.Add(shortCode)
		http.Redirect(w, r, url.Original_url, http.StatusFound)
	}
}

func GetStats(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")
	if shortCode == "" {
		http.Error(w, "short_code is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	var stats models.StatsResponse
	query := `
        SELECT short_code, original_url, created_at, click_count, last_accessed_at
        FROM urls
        WHERE short_code = $1;
    `
	err := db.DB.QueryRow(ctx, query, shortCode).Scan(
		&stats.ShortCode,
		&stats.OriginalURL,
		&stats.CreatedAt,
		&stats.ClickCount,
		&stats.LastAccessedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "url not found", http.StatusNotFound)
			return
		}
		log.Println("stats query:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := helpers.JSONencoder(w, &stats, http.StatusOK); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func ReadyCheck(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if err := db.DB.Ping(ctx); err != nil {
		log.Println("ReadyCheck: db ping failed:", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("db not ready"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
