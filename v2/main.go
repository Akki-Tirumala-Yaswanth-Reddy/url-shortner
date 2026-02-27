package main

import (
    "log"
    "net/http"
    "url_shortner/m/v2/db"
    "url_shortner/m/v2/handlers"
    "url_shortner/m/v2/middleware"
)

func main() {
    if err := db.InitDB(); err != nil {
        log.Fatalf("failed to initialize db: %v", err)
    }
    defer db.DB.Close()

    mux := http.NewServeMux()
    mux.Handle("POST /create", middleware.LoggingMiddleware(http.HandlerFunc(handlers.CreateShortUrl)))
    mux.Handle("GET /redirect/{short_code}", middleware.LoggingMiddleware(http.HandlerFunc(handlers.Redirect)))

    log.Println("Server running at: localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}