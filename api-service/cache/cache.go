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
	return Client.Get(ctx, cacheKey).Result()
}

func SetURL(shortCode, longURL string, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("url:%s", shortCode)
	return Client.Set(ctx, cacheKey, longURL, ttl).Err()
}

func IncrementClickCounter(shortCode string) error {
	counterKey := fmt.Sprintf("clicks:%s", shortCode)
	return Client.Incr(ctx, counterKey).Err()
}

func GetRateLimit(key string) (int, error) {
	rateLimitKey := fmt.Sprintf("rate_limit:%s", key)
	return Client.Get(ctx, rateLimitKey).Int()
}

func IncrementRateLimit(key string, window int) error {
	rateLimitKey := fmt.Sprintf("rate_limit:%s", key)
	pipe := Client.Pipeline()
	pipe.Incr(ctx, rateLimitKey)
	pipe.Expire(ctx, rateLimitKey, time.Duration(window)*time.Second)
	_, err := pipe.Exec(ctx)
	return err
}

func CheckHealth() bool {
	return Client.Ping(ctx).Err() == nil
}
