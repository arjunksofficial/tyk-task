package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/arjunksofficial/tyk-task/internal/rediscli"
	"github.com/arjunksofficial/tyk-task/internal/token/models"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	GetToken(ctx context.Context, token string) (models.TokenData, error)
	StoreToken(ctx context.Context, token models.TokenData) error
	DeleteToken(ctx context.Context, token string) error

	IncrementRateLimit(ctx context.Context, token string) (int64, error)
}

type service struct {
	redisClient *redis.Client
}

func New() Service {
	return &service{
		redisClient: rediscli.GetRedisClient(),
	}
}
func (s *service) GetToken(ctx context.Context, token string) (models.TokenData, error) {
	key := "token:" + token

	// Fetch the token data from Redis
	data, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return models.TokenData{}, redis.Nil
		}
		return models.TokenData{}, err
	}

	var tokenData models.TokenData
	if err := json.Unmarshal([]byte(data), &tokenData); err != nil {
		return models.TokenData{}, err
	}

	return tokenData, nil
}

func (s *service) StoreToken(ctx context.Context, token models.TokenData) error {
	key := "token:" + token.APIKey
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	// Store the token data in Redis
	return s.redisClient.Set(ctx, key, data, 0).Err()
}

func (s *service) DeleteToken(ctx context.Context, token string) error {
	key := "token:" + token
	// Delete the token data from Redis
	return s.redisClient.Del(ctx, key).Err()
}

func (s *service) IncrementRateLimit(ctx context.Context, token string) (int64, error) {
	// Fixed window key: token:<api_key>:rate:<YYYYMMDDHHMM>
	window := time.Now().UTC().Format("200601021504")
	key := "rate_limit:" + token + ":" + window

	// Increment the rate limit counter
	count, err := s.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Set an expiration time for the rate limit key if it is the first increment
	if count == 1 {
		s.redisClient.Expire(ctx, key, time.Minute) // 1 minute
	}

	return count, nil
}
