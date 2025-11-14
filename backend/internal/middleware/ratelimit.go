package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lingopaste/backend/internal/db"
	"github.com/lingopaste/backend/internal/utils"
)

type RateLimiter struct {
	db *db.DynamoDB
}

func NewRateLimiter(db *db.DynamoDB) *RateLimiter {
	return &RateLimiter{db: db}
}

func (rl *RateLimiter) CheckRateLimit(accountID string, isPaid bool, ip string) error {
	ctx := context.Background()
	today := time.Now().Format("2006-01-02")

	if isPaid {
		return rl.checkAccountLimit(ctx, accountID, today, 1000)
	}

	if accountID != "" {
		if err := rl.checkAccountLimit(ctx, accountID, today, 5); err != nil {
			return err
		}
		return rl.checkIPLimit(ctx, ip, today, 50)
	}

	return rl.checkIPLimit(ctx, ip, today, 5)
}

func (rl *RateLimiter) checkAccountLimit(ctx context.Context, accountID, date string, limit int) error {
	count, err := rl.db.IncrementRateLimit(ctx, accountID, date, "account")
	if err != nil {
		log.Printf("Error incrementing account rate limit: %v", err)
		return fmt.Errorf("rate limit check failed")
	}

	if count > limit {
		return fmt.Errorf("account rate limit exceeded: %d/%d pastes today", count-1, limit)
	}

	return nil
}

func (rl *RateLimiter) checkIPLimit(ctx context.Context, ip, date string, limit int) error {
	ipHash := utils.HashIP(ip)
	count, err := rl.db.IncrementRateLimit(ctx, ipHash, date, "ip")
	if err != nil {
		log.Printf("Error incrementing IP rate limit: %v", err)
		return fmt.Errorf("rate limit check failed")
	}

	if count > limit {
		return fmt.Errorf("IP rate limit exceeded: %d/%d pastes today", count-1, limit)
	}

	return nil
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only apply rate limiting to paste creation
		if r.URL.Path == "/api/pastes" && r.Method == "POST" {
			ip := GetIPFromContext(r.Context())
			// TODO: Get account info from JWT token
			accountID := ""
			isPaid := false

			if err := rl.CheckRateLimit(accountID, isPaid, ip); err != nil {
				http.Error(w, err.Error(), http.StatusTooManyRequests)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
