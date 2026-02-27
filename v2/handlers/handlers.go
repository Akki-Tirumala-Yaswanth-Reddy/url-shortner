package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"url_shortner/m/v2/db"
	"url_shortner/m/v2/helpers"
	"url_shortner/m/v2/models"

	"github.com/jackc/pgx/v5/pgconn"
)

var counter = 0
var mux sync.RWMutex

func CreateShortUrl(w http.ResponseWriter, r *http.Request) {
	var url models.Url
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

	mux.Lock()
	counter++
	mux.Unlock()
	shortUrl := "localhost:8080/redirect/" + strconv.Itoa(counter)
	res.Url = shortUrl
	url.Short_code = strconv.Itoa(counter)

	query := `
		INSERT INTO urls (user_name, short_code, original_url)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	err := db.DB.QueryRow(context.Background(), query, req.User, url.Short_code, req.Url).Scan(&res.Id)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		log.Println(errors.Is(err, pgErr))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := helpers.JSONencoder(w, &res, http.StatusOK); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	var url models.Url
	short_code := r.PathValue("short_code")
	_, err := strconv.Atoi(short_code)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	query := `
		SELECT user_name, short_code, original_url
		FROM urls
		WHERE short_code=$1;
	`
	err = db.DB.QueryRow(context.Background(), query, short_code).Scan(&url.Username, &url.Short_code, &url.Original_url)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "url not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, url.Original_url, http.StatusFound)
}
