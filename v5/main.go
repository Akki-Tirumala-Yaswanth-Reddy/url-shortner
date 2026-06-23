package main

import (
	"log"
	"net/http"
	"os"
	"url_shortner/v5/analytics"
	"url_shortner/v5/db"
	"url_shortner/v5/handlers"
	"url_shortner/v5/middleware"
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

	ana := analytics.NewAnalyticsObj()
	go analytics.StartAnalytics(ana)

	mux := http.NewServeMux()
	mux.Handle("POST /create", middleware.LoggingMiddleware(http.HandlerFunc(handlers.CreateShortUrl)))
	mux.Handle("GET /redirect/{short_code}", middleware.LoggingMiddleware(http.HandlerFunc(handlers.Redirect(ana))))
	mux.Handle("GET /stats/{short_code}", middleware.LoggingMiddleware(http.HandlerFunc(handlers.GetStats)))
	mux.Handle("GET /healthCheck", http.HandlerFunc(handlers.HealthCheck))
	mux.Handle("GET /readyCheck", http.HandlerFunc(handlers.ReadyCheck))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	address := ":" + port
	log.Println("Server running at: localhost:" + port)
	log.Fatal(http.ListenAndServe(address, mux))
}
