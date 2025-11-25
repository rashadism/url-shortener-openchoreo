package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	Client *redis.Client
	ctx    = context.Background()
)

func Init(redisURL string) error {
	Client = redis.NewClient(&redis.Options{
		Addr: redisURL,
		DB:   0,
	})

	if err := Client.Ping(ctx).Err(); err != nil {
		return err
	}

	log.Println("Redis connected successfully")
	return nil
}

func Close() {
	if Client != nil {
		Client.Close()
	}
}

func GetURL(shortCode string) (string, error) {
	cacheKey := fmt.Sprintf("url:%s", shortCode)
	result, err := Client.Get(ctx, cacheKey).Result()
	if err == nil {
		log.Printf("[Redis] Cache HIT for key: %s", cacheKey)
	}
	return result, err
}

func SetURL(shortCode, longURL string, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("url:%s", shortCode)
	err := Client.Set(ctx, cacheKey, longURL, ttl).Err()
	if err == nil {
		log.Printf("[Redis] Saved to cache: %s (TTL: %v)", cacheKey, ttl)
	} else {
		log.Printf("[Redis] Failed to save to cache: %s - %v", cacheKey, err)
	}
	return err
}

func DeleteURL(shortCode string) error {
	cacheKey := fmt.Sprintf("url:%s", shortCode)
	err := Client.Del(ctx, cacheKey).Err()
	if err == nil {
		log.Printf("[Redis] Deleted from cache: %s", cacheKey)
	} else {
		log.Printf("[Redis] Failed to delete from cache: %s - %v", cacheKey, err)
	}
	return err
}

func IncrementClickCounter(shortCode string) error {
	counterKey := fmt.Sprintf("clicks:%s", shortCode)
	result, err := Client.Incr(ctx, counterKey).Result()
	if err == nil {
		log.Printf("[Redis] Incremented click counter for %s to %d", counterKey, result)
	} else {
		log.Printf("[Redis] Failed to increment click counter: %s - %v", counterKey, err)
	}
	return err
}

func GetRateLimit(key string) (int, error) {
	rateLimitKey := fmt.Sprintf("rate_limit:%s", key)
	result, err := Client.Get(ctx, rateLimitKey).Int()
	if err == nil {
		log.Printf("[Redis] Rate limit check for %s: %d requests", rateLimitKey, result)
	}
	return result, err
}

func IncrementRateLimit(key string, window int) error {
	rateLimitKey := fmt.Sprintf("rate_limit:%s", key)
	pipe := Client.Pipeline()
	incrCmd := pipe.Incr(ctx, rateLimitKey)
	pipe.Expire(ctx, rateLimitKey, time.Duration(window)*time.Second)
	_, err := pipe.Exec(ctx)
	if err == nil {
		count := incrCmd.Val()
		log.Printf("[Redis] Incremented rate limit for %s to %d (window: %ds)", rateLimitKey, count, window)
	} else {
		log.Printf("[Redis] Failed to increment rate limit: %s - %v", rateLimitKey, err)
	}
	return err
}

func CheckHealth() bool {
	return Client.Ping(ctx).Err() == nil
}
