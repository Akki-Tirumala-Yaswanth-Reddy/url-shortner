package main

import (
	"log"
	"net/http"
	"os"
	"url_shortner/v3/db"
	"url_shortner/v3/handlers"
	"url_shortner/v3/middleware"
)

func main() {
	if err := db.InitDB(); err != nil {
		log.Fatalf("failed to initialize db: %v", err)
	}
	defer db.DB.Close()

	if err := db.InitRedis(); err != nil {
		log.Fatalf("failed to initialize redis: %v", err)
	}
	defer db.RDB.Close()

	mux := http.NewServeMux()
	mux.Handle("POST /create", middleware.LoggingMiddleware(http.HandlerFunc(handlers.CreateShortUrl)))
	mux.Handle("GET /redirect/{short_code}", middleware.LoggingMiddleware(http.HandlerFunc(handlers.Redirect)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	address := ":" + port
	log.Println("Server running at: localhost:8080")
	log.Fatal(http.ListenAndServe(address, mux))
}
