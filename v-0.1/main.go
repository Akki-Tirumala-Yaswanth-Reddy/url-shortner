package main

import (
	"net/http"
	"strconv"
	"sync"
)

type urlModel struct {
	user     string
	longUrl string
}

type RequestJSONmodel struct {
	User string `json:"user"`
	Url  string `json:"url"`
}

type ResponseJSONmodel struct {
	Url string `json:"url"`
}

var cache = make(map[int]urlModel)
var counter int = 0
var mux sync.RWMutex

func main() {
	mux := http.NewServeMux()
	mux.Handle("POST /create", LoggingMiddleware(http.HandlerFunc(CreateShortUrl)))
	mux.Handle("/redirect/{url}", LoggingMiddleware(http.HandlerFunc(Redirect)))
	http.ListenAndServe(":8080", mux)
}

func CreateShortUrl(w http.ResponseWriter, r *http.Request) {
	var req RequestJSONmodel
	if err := JSONdecoder(r, &req); err != nil {
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

	counter++
	shortUrl := "localhost:8080/redirect/" + strconv.Itoa(counter)

	mux.Lock()
	cache[counter] = urlModel{req.User, req.Url}
	mux.Unlock()

	res := ResponseJSONmodel{shortUrl}
	if err := JSONencoder(w, &res, http.StatusOK); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("url"))
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
	}
	mux.RLock()
	val, ok := cache[id]
	if !ok {
		http.Error(w, "url not found", http.StatusNotFound)
		mux.RUnlock()
		return
	}
	mux.RUnlock()
	http.Redirect(w, r, val.longUrl, http.StatusFound)
}
