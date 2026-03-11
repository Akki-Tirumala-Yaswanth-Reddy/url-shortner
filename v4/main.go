package main

import (
	"log"
	"net/http"
	"os"
	"url_shortner/v4/db"
	"url_shortner/v4/handlers"
	"url_shortner/v4/middleware"
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
	mux.Handle("GET /api/v1/links/{short_code}/stats", middleware.LoggingMiddleware(http.HandlerFunc(handlers.GetStats)))
	mux.Handle("GET /healthz", http.HandlerFunc(handlers.Healthz))
	mux.Handle("GET /readyz", http.HandlerFunc(handlers.Readyz))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	address := ":" + port
	log.Println("Server running at: localhost:" + port)
	log.Fatal(http.ListenAndServe(address, mux))
}
