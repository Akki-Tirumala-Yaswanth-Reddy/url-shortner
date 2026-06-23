package middleware

import (
	"context"
	"net/http"
	"time"
	"url_shortner/v5/db"
)

func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler{
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ip := r.RemoteAddr
			ctx := context.Background()
			key := "ratelimit:" + ip

			count, err := db.RDB.Incr(ctx, key).Result()
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			if count == 1 {
				db.RDB.Expire(ctx, key, window)
			}

			if count > int64(limit) {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
