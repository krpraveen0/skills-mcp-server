package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
)

// Redis wraps the go-redis client with typed helpers.
type Redis struct {
	client *redis.Client
}

// New creates a new Redis client from a URL.
func New(redisURL, password string) (*Redis, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	if password != "" {
		opt.Password = password
	}

	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	return &Redis{client: client}, nil
}

// Get deserializes a cached JSON value into dest.
func (r *Redis) Get(ctx context.Context, key string, dest any) error {
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(val, dest)
}

// Set serializes src to JSON and stores it with the given TTL.
func (r *Redis) Set(ctx context.Context, key string, src any, ttl time.Duration) error {
	data, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("marshal cache value: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a key from the cache.
func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// DeletePattern removes all keys matching a glob pattern.
func (r *Redis) DeletePattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		if cursor == 0 {
			break
		}
	}
	return nil
}

// Close closes the Redis connection.
func (r *Redis) Close() error {
	return r.client.Close()
}
