package url

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"time"
)

var (
	maxRate = 10
	window  = time.Minute
	// defaultExpiry = 30 * time.Minute
)

type Repository struct {
	redisClient *redis.Client
	ctx         context.Context
}

func NewRedisClient(dbNo int) *Repository {
	return &Repository{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     os.Getenv("DB_ADDR"),
			Username: os.Getenv("DB_PASS"),
			DB:       dbNo,
		}),
		ctx: context.Background(), // todo context
	}
}

func (r *Repository) Get(key string) (string, error) {
	value, err := r.redisClient.Get(r.ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return "", err // todo return not found
	} else if err != nil {
		return "", err // todo return 500
	}

	return value, nil
}

func (r *Repository) Set(key string, value string, expiry time.Duration) error {
	err := r.redisClient.Set(r.ctx, key, value, expiry).Err() // todo check expiry
	if err != nil {
		return err // todo return 500
	}
	return nil
}

// isAllowed checks if the given IP address is allowed to make requests based on a rate-limiting mechanism.
// It adds the current time to a sorted set in Redis, representing request timestamps.
// Then, it removes old entries from the set to maintain a sliding time window.
// The function then calculates the remaining rate of requests within the window for the IP address.
// If the remaining rate is greater than 0, it returns true, indicating that the IP is allowed to make more requests.
// Otherwise, it returns false, indicating that the IP has exceeded the rate limit.
func (r *Repository) isAllowed(IP string) bool {
	key := fmt.Sprintf("ratelimit:%s", IP)

	currentTime := time.Now().Unix()

	// Add the current time to the sorted set
	if _, err := r.redisClient.ZAdd(r.ctx, key, redis.Z{
		Score:  float64(currentTime),
		Member: currentTime,
	}).Result(); err != nil {
		panic(err)
	}

	// Remove old entries from the sorted set
	if _, err := r.redisClient.ZRemRangeByScore(
		r.ctx,
		key,
		"-inf", fmt.Sprintf("(%d", currentTime-int64(window.Seconds()))).Result(); err != nil {
		panic(err)
	}
	// Get the number of requests made within the time window
	remainingRate, err := r.calculateRemainingRate(IP)
	if err != nil {
		panic(err)
	}

	// Check if the number of requests exceeds the maxRate
	return remainingRate > 0
}

func (r *Repository) TTL(IP string) time.Duration {
	ttl, _ := r.redisClient.TTL(r.ctx, IP).Result()

	return ttl / time.Nanosecond / time.Minute
}

// calculateRemainingRate calculates the remaining rate of requests allowed for the given IP address.
// It retrieves the count of requests made within the time window from Redis.
// If no requests are found, it assumes one less than the maximum rate.
// If an error occurs, it returns 0 and the error.
// Otherwise, it returns the remaining rate and nil error.
func (r *Repository) calculateRemainingRate(IP string) (int, error) {
	key := fmt.Sprintf("ratelimit:%s", IP)
	// Get the number of requests made within the time window
	count, err := r.redisClient.ZCard(r.ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return maxRate - 1, nil
	} else if err != nil {
		return 0, err
	}
	return maxRate - int(count), nil
}
