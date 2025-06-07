package rediscli

import (
	"context"

	"github.com/arjunksofficial/tyk-task/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient() (*redis.Client, error) {
	// Load configuration
	cfg := config.GetConfig()
	redisConfig := cfg.GetRedisConfig()

	// Create a new Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Host + ":" + redisConfig.Port,
		DB:       redisConfig.DB,
		Password: redisConfig.Password,
	})

	// Test the connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}

// GetRedisClient returns a singleton Redis client instance
var redisClient *redis.Client

func GetRedisClient() *redis.Client {
	if redisClient == nil {
		client, err := NewRedisClient()
		if err != nil {
			panic("Failed to create Redis client: " + err.Error())
		}
		redisClient = client
	}
	return redisClient
}

// CloseRedisClient closes the Redis client connection
func CloseRedisClient() error {
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			return err
		}
		redisClient = nil
	}
	return nil
}
