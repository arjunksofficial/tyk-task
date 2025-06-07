package auth

import (
	"context"
	"net/http"

	tokenservice "github.com/arjunksofficial/tyk-task/internal/token/services"
	"github.com/redis/go-redis/v9"
)

type AuthMiddleware struct {
	TokenService tokenservice.Service
}

// NewAuthMiddleware creates a new instance of AuthMiddleware
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		TokenService: tokenservice.New(),
	}
}

// AuthMiddleware is a middleware function that checks if the request has a valid API key
func (a *AuthMiddleware) AuthMiddleware(next http.Handler) http.Handler {
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
		token, err := a.TokenService.GetToken(r.Context(), apiKey)
		if err != nil {
			if err == redis.Nil {
				http.Error(w, "Unauthorized: Invalid API key", http.StatusUnauthorized)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		// Check if the token is valid
		if !token.IsValid() {
			http.Error(w, "Unauthorized: Invalid API key", http.StatusUnauthorized)
			return
		}
		// Set the token in the request context for further processing
		ctx := r.Context()
		ctx = context.WithValue(ctx, "token", token)
		r = r.WithContext(ctx)

		// If the API key is valid, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}
