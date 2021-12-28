package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kwitsch/ArpRedisCollector/config"
	"github.com/kwitsch/ArpRedisCollector/models"
)

const noTTL time.Duration = time.Duration(0)

type Client struct {
	cfg    *config.RedisConfig
	client *redis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new redis client
func New(cfg *config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            cfg.Address,
		Username:        cfg.Username,
		Password:        cfg.Password,
		DB:              cfg.Database,
		MaxRetries:      cfg.ConnectionAttempts,
		MaxRetryBackoff: time.Duration(cfg.ConnectionCooldown),
	})
	ctx, cancel := context.WithCancel(context.Background())

	_, err := rdb.Ping(ctx).Result()
	if err == nil {
		res := &Client{
			cfg:    cfg,
			client: rdb,
			ctx:    ctx,
			cancel: cancel,
		}
		return res, nil
	}
	cancel()
	return nil, err
}

// Close discards the redis client
func (c *Client) Close() {
	c.cancel()
}

// Publish stores a MACEntry in redis
func (c *Client) Publish(cm *models.CacheMessage) {
	c.client.Set(c.ctx, cm.Mac.String(), cm.IP.String(), noTTL)
}
