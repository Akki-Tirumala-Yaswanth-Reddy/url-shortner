package expiration

import (
	"context"
	"log"
	"time"
	"url_shortner/v5/db"
)

func cleanExpiry() {
	ctx := context.Background()

	query := `
		DELETE FROM urls
		WHERE created_at < NOW() - INTERVAL '2 months'
	`

	if _, err := db.DB.Exec(ctx, query); err != nil {
		log.Printf("error in cleaning expired links: %v\n", err)
		return
	}
	log.Println("cleaned expired links")
}

func StartCleanExpiry() {
	cleanExpiry()

	timer := time.NewTicker(time.Hour * 12)
	defer timer.Stop()

	for range timer.C {
		cleanExpiry()
	}
}
