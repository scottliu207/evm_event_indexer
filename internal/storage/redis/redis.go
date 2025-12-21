package redis

import (
	"context"
	"evm_event_indexer/internal/config"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisManager struct {
	pool map[string]*redis.Client
	mu   sync.Mutex
}

func NewRedisManager() *RedisManager {
	return &RedisManager{
		pool: make(map[string]*redis.Client),
	}
}

func (m *RedisManager) GetRedis(name string) (*redis.Client, error) {
	db, ok := m.pool[name]
	if !ok {
		return nil, fmt.Errorf("redis %s not initialized", name)
	}

	return db, nil
}

func (m *RedisManager) InitRedis(name string, dbCnf config.Redis) error {

	r := redis.NewClient(
		&redis.Options{
			ReadTimeout:     dbCnf.ReadTimeout,
			WriteTimeout:    dbCnf.WriteTimeout,
			Addr:            dbCnf.Host,
			MaxRetries:      dbCnf.MaxRetries,
			DialTimeout:     dbCnf.DialTimeout,
			PoolSize:        dbCnf.PoolSize,
			PoolTimeout:     dbCnf.PoolTimeout,
			ConnMaxIdleTime: dbCnf.IdleTimeout,
			ConnMaxLifetime: dbCnf.IdleCheckFrequency,
			DB:              dbCnf.DB,
		},
	)

	for i := 0; i < config.Get().Redis.Retry; i++ {
		err := func() error {
			ctxTimeout, cancel := context.WithTimeout(
				context.Background(),
				config.Get().Redis.PingTimeout,
			)
			defer cancel()

			if err := r.Ping(ctxTimeout).Err(); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			slog.Warn("redis initialize failure, failed to connect to redis",
				slog.Any("redis Name", name),
				slog.Any("error", err),
				slog.Any("waiting for retry", config.Get().Redis.WaitDuration.String()),
			)
			time.Sleep(config.Get().Redis.WaitDuration)
			continue
		}

		m.mu.Lock()
		defer m.mu.Unlock()
		m.pool[name] = r
		return nil
	}
	r.Close()
	return fmt.Errorf("redis initialize failure, retry limit exceed")
}

func (m *RedisManager) Shutdown(name string) error {
	db, err := m.GetRedis(name)
	if err != nil {
		return fmt.Errorf("failed to get redis: %w", err)
	}
	db.Close()
	delete(m.pool, name)
	return nil
}
