package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"
	"url_shortner/v4/db"
	"url_shortner/v4/helpers"
	"url_shortner/v4/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

	// Cache the new short code -> original URL mapping in Redis
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

func Redirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")
	if shortCode == "" {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Atomic: increment click_count, set last_accessed_at, return original_url
	var originalURL string
	query := `
		UPDATE urls
		SET click_count = click_count + 1,
		    last_accessed_at = now()
		WHERE short_code = $1
		RETURNING original_url;
	`
	if err := db.DB.QueryRow(ctx, query, shortCode).Scan(&originalURL); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "url not found", http.StatusNotFound)
			return
		}
		log.Println("redirect query:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Update Redis cache with the original URL
	if err := db.RDB.Set(ctx, "url:"+shortCode, originalURL, 24*time.Hour).Err(); err != nil {
		log.Println("redis set:", err)
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
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

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func Readyz(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if err := db.DB.Ping(ctx); err != nil {
		log.Println("readyz: db ping failed:", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("db not ready"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
