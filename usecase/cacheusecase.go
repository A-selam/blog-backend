package usecase

import (
	"context"
	"fmt"
	"time"
	"log" 

	"blog-backend/domain" 
	"github.com/go-redis/redis/v8"
)

type cacheUseCase struct {
	cacheRepo      domain.ICacheRepository
	contextTimeout time.Duration
	redisClient *redis.Client 
}

func NewCacheUseCase(
	cacheRepo domain.ICacheRepository,
	redisClient *redis.Client, 
	timeout time.Duration,
) domain.ICacheUseCase {
	return &cacheUseCase{
		cacheRepo:      cacheRepo,
		redisClient:    redisClient,
		contextTimeout: timeout,
	}
}

func (uc *cacheUseCase) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	return uc.cacheRepo.Set(ctx, key, value, expiration)
}

func (uc *cacheUseCase) Get(ctx context.Context, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	return uc.cacheRepo.Get(ctx, key)
}

func (uc *cacheUseCase) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	return uc.cacheRepo.Delete(ctx, key)
}

func (uc *cacheUseCase) InvalidatePrefix(ctx context.Context, prefix string) error {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	var cursor uint64
	var keys []string
	var err error

	for {
		keys, cursor, err = uc.redisClient.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			log.Printf("Error scanning Redis keys with prefix %s: %v", prefix, err)
			return fmt.Errorf("failed to scan keys: %w", err)
		}

		if len(keys) > 0 {
			if err := uc.redisClient.Del(ctx, keys...).Err(); err != nil {
				log.Printf("Error deleting keys from Redis: %v", err)
				return fmt.Errorf("failed to delete keys: %w", err)
			}
		}

		if cursor == 0 { 
			break
		}
	}
	log.Printf("Successfully invalidated cache for prefix: %s", prefix)
	return nil
}
