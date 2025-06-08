package ratelimit

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/arjunksofficial/tyk-task/internal/token/models"
	tokenservice "github.com/arjunksofficial/tyk-task/internal/token/services"
	"github.com/go-redis/redis"
)

// RateLimitMiddleware is a middleware that limits the number of requests per API key
type RateLimitMiddleware struct {
	TokenService tokenservice.Service
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware
func NewRateLimitMiddleware() *RateLimitMiddleware {
	return &RateLimitMiddleware{
		TokenService: tokenservice.New(),
	}
}

// RateLimitHandler is the middleware handler that checks the rate limit
func (rl *RateLimitMiddleware) RateLimitHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("Authorization")
		// Check if the API key is present
		if apiKey == "" {
			http.Error(w, "Unauthorized: API key is missing", http.StatusUnauthorized)
			return
		}
		// remove Bearer prefix if present
		if len(apiKey) > 7 && apiKey[:7] == "Bearer " {
			apiKey = apiKey[7:]
		}
		token, err := rl.TokenService.GetToken(r.Context(), apiKey)
		if err != nil {
			if err == redis.Nil {
				http.Error(w, "Unauthorized: Invalid API key", http.StatusUnauthorized)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		expiryTime, err := time.Parse(time.RFC3339, token.ExpiresAt)
		if err != nil {
			http.Error(w, "Internal Server Error: Invalid token expiry format", http.StatusInternalServerError)
			return
		}
		// Check if the token is expired
		if time.Now().UTC().After(expiryTime) {
			http.Error(w, "Unauthorized: Invalid API key", http.StatusUnauthorized)
			return
		}

		// Check if route is allowed
		allowed := false
		for _, pattern := range token.AllowedRoutes {
			if matchPath(pattern, r.URL.Path) {
				allowed = true
				break
			}
		}

		if !allowed {
			http.Error(w, "Route not allowed for this token", http.StatusForbidden)
			return
		}

		// Fixed window key: token:<api_key>:rate:<YYYYMMDDHHMM>
		window := time.Now().UTC().Format("200601021504")
		rateKey := "ratelimit:" + apiKey + ":" + window

		// Increment the rate limit count in Redis
		count, err := rl.TokenService.IncrementRateLimit(r.Context(), rateKey)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if int(count) > token.RateLimit {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, models.TokenContextKey, token)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// Simple wildcard matcher: "/api/v1/users/*" matches "/api/v1/users/123"
func matchPath(pattern, path string) bool {
	if strings.HasSuffix(pattern, "/*") {
		return strings.HasPrefix(path, strings.TrimSuffix(pattern, "*"))
	}
	return pattern == path
}
