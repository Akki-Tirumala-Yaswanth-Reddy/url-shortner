package main

import (
	"log"
	"net/http"
	"os"
	"time"
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
	// run a go routine in the background to update the visits every 1 minute
	go analytics.StartAnalytics(ana)

	mux := http.NewServeMux()
	// middleware.RateLimit returns func which takes a func(http.Handler) as input
	mux.Handle("POST /create", middleware.RateLimit(5, time.Minute)(middleware.LoggingMiddleware(http.HandlerFunc(handlers.CreateShortUrl))))
	mux.Handle("GET /redirect/{short_code}", middleware.RateLimit(100, time.Minute)(middleware.LoggingMiddleware(http.HandlerFunc(handlers.Redirect(ana)))))
	mux.Handle("GET /stats/{short_code}", middleware.RateLimit(100, time.Minute)(middleware.LoggingMiddleware(http.HandlerFunc(handlers.GetStats))))
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
