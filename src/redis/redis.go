package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/irai/arp"
	"github.com/kwitsch/ArpRedisCollector/config"
)

type Client struct {
	cfg    *config.RedisConfig
	client *redis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

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

func (c *Client) Close() {
	c.cancel()
}

func (c *Client) Publish(entry *arp.MACEntry) {
	if entry.Online {
		c.client.Set(c.ctx, entry.MAC.String(), entry.IP().String(), c.cfg.TTL)
	}
}
