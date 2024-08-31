package redis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tizzhh/micro-banking/internal/config"
	"github.com/tizzhh/micro-banking/internal/storage"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

type Cache struct {
	keyTTL time.Duration
	rdb    *redis.Client
	log    *slog.Logger
}

var cacheInstance *Cache
var once sync.Once

func Get(log *slog.Logger) (*Cache, error) {
	var err error
	once.Do(func() {
		cacheInstance, err = New(log)
	})
	return cacheInstance, err
}

func New(log *slog.Logger) (*Cache, error) {
	cfg := config.Get()

	return &Cache{
		log: log,
		rdb: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
		}),
	}, nil
}

func (c *Cache) MustPing(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Cache) GetCurrencyRate(ctx context.Context, currencyCode string) (float32, error) {
	const caller = "storage.redis.GetCurrencyRate"

	log := sl.AddCaller(c.log, caller)

	log.Info("getting currency rate", slog.String("currency", currencyCode))

	strRate, err := c.rdb.Get(ctx, currencyCode).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			log.Warn("currency key not found", slog.String("currency", currencyCode))
			return 0, fmt.Errorf("%s: %w", caller, storage.ErrCurrencyKeyNotFound)
		}
		log.Error("failed to get currency rate", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	rate, err := strconv.ParseFloat(strRate, 32)
	if err != nil {
		log.Error("failed to convert rate to float", slog.String("currency", currencyCode), slog.String("value", strRate))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	return float32(rate), nil
}

func (c *Cache) SetCurrencyRate(ctx context.Context, currencyCode string, rate float32) error {
	const caller = "storage.redis.SetCurrencyRate"

	log := sl.AddCaller(c.log, caller)

	if err := c.rdb.Set(ctx, currencyCode, rate, c.keyTTL).Err(); err != nil {
		log.Error("failed to set key", sl.Error(err))
		return fmt.Errorf("%s: %w", caller, err)
	}

	return nil
}
