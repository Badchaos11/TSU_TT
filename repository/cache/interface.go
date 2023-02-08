package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type ICacheRepository interface {
	AddToCache(ctx context.Context, key, value string) error
	GetUsersFromCache(ctx context.Context, key string) (string, error)
	DeleteValue(ctx context.Context, key string) error
	ClearCache(ctx context.Context) error
}

type kvRepo struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCacheClient(ctx context.Context, url string) (ICacheRepository, error) {
	fmt.Println("url redis ", url)
	rdb := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		logrus.Errorf("Error connecting to redis: %v", err)
		return nil, err
	}

	return &kvRepo{client: rdb, ttl: time.Hour * 1}, nil
}
