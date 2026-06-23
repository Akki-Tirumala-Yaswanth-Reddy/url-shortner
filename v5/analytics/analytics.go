package analytics

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
	"url_shortner/v5/db"

	"github.com/jackc/pgx/v5"
)

type AnalyticsObj struct {
	count map[string]int
	mu    sync.Mutex
}

func NewAnalyticsObj() *AnalyticsObj {
	return &AnalyticsObj{count: make(map[string]int)}
}

func (ana *AnalyticsObj) Add(shortCode string) {
	ana.mu.Lock()
	defer ana.mu.Unlock()
	ana.count[shortCode]++
}

func (ana *AnalyticsObj) update() {
	ctx := context.Background()

	ana.mu.Lock()
	local := ana.count
	if len(local) == 0 {
		ana.mu.Unlock()
		return
	}
	ana.count = make(map[string]int)
	ana.mu.Unlock()

	query := `
        UPDATE urls
        SET click_count = click_count + $1,
            last_accessed_at = now()
        WHERE short_code = $2;
    `

	for shortCode, value := range local {
		if _, err := db.DB.Exec(ctx, query, value, shortCode); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				log.Printf("short code not found: %s\n", shortCode)
				continue
			}
			log.Printf("update failed for %s: %v\n", shortCode, err)
			continue
		}
	}
	log.Println("click count updated")
}

func StartAnalytics(ana *AnalyticsObj) {
	timer := time.NewTicker(time.Minute)
	defer timer.Stop()

	for range timer.C {
		ana.update()
	}
}
