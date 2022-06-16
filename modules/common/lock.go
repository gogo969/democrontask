package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

var (
	ctx = context.Background()
	// The maximum duration to lock a key, Default: 10s
	LockTimeout time.Duration = 10 * time.Second
	// The maximum duration to wait to get the lock, Default: 0s, do not wait
	WaitTimeout time.Duration
	// The maximum wait retry time to get the lock again, Default: 100ms
	WaitRetry time.Duration = 100 * time.Millisecond
)

const (
	defaultRedisKeyPrefix = "rlock:"
)

func Lock(pool *redis.ClusterClient, key string, ttl time.Duration) error {

	val := fmt.Sprintf("%s%s", defaultRedisKeyPrefix, key)
	ok, err := pool.SetNX(ctx, val, "1", ttl).Result()
	if err != nil {
		return fmt.Errorf("get lock failed, reason:%s", err.Error())
	}

	if !ok {
		return errors.New("get lock failed")
	}

	return nil
}

func LockExpireAt(pool *redis.ClusterClient, key string, tm time.Time) error {

	val := fmt.Sprintf("%s%s", defaultRedisKeyPrefix, key)

	ok, err := pool.SetNX(ctx, val, "1", 2*time.Hour).Result()
	if err != nil {
		return fmt.Errorf("get lock failed, reason:%s", err.Error())
	}

	if !ok {
		return errors.New("get lock failed")
	}

	return pool.ExpireAt(ctx, val, tm).Err()
}

func Unlock(pool *redis.ClusterClient, key string) {

	val := fmt.Sprintf("%s%s", defaultRedisKeyPrefix, key)
	pool.Unlink(ctx, val)
}
