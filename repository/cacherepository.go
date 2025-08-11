package repository

import (
	"context"
	"time"

	"blog-backend/domain" 
	"github.com/redis/go-redis/v9"  
)


type cacheRepository struct {
	redisClient *redis.Client
}

func NewCacheRepository(client *redis.Client) domain.ICacheRepository {
	return &cacheRepository{
		redisClient: client,
	}
}


func (r *cacheRepository) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return r.redisClient.Set(ctx, key, value, expiration).Err()
}


func (r *cacheRepository) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.redisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (r *cacheRepository) Delete(ctx context.Context, key string) error {

	return r.redisClient.Del(ctx, key).Err()
}

func (r *cacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
