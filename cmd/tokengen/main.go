package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/arjunksofficial/tyk-task/internal/config"
	"github.com/arjunksofficial/tyk-task/internal/rediscli"
	"github.com/arjunksofficial/tyk-task/internal/token/models"
	"github.com/google/uuid"
)

func main() {
	ctx := context.Background()

	config.GetConfig() // Connect to Redis
	redisCli := rediscli.GetRedisClient()
	defer func() {
		if err := rediscli.CloseRedisClient(); err != nil {
			log.Fatalf("Failed to close Redis client: %v", err)
		}
	}()

	// Generate a UUID-based API Key
	apiKey := uuid.New().String()

	// Define the token metadata
	token := models.TokenData{
		APIKey:        apiKey,
		RateLimit:     5,
		ExpiresAt:     time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
		AllowedRoutes: []string{"/api/v1/users/*", "/api/v1/products/*", "/api/v1/orders/*"},
	}

	// Convert to JSON for storage
	jsonData, err := json.Marshal(token)
	if err != nil {
		log.Fatalf("Failed to marshal token: %v", err)
	}

	// Store in Redis under the key: "token:<api_key>"
	key := fmt.Sprintf("token:%s", apiKey)
	err = redisCli.Set(ctx, key, jsonData, 0).Err() // No TTL at Redis level (we handle expiry in code)
	if err != nil {
		log.Fatalf("Failed to store token in Redis: %v", err)
	}

	fmt.Println("✅ Token generated and stored in Redis:")
	fmt.Printf("API Key: %s\n", apiKey)
}
