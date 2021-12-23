package config

import (
	"fmt"
	"time"

	. "github.com/kwitsch/go-dockerutils/config"
)

type Config struct {
	Redis   RedisConfig    `koanf:"redis"`
	Verbose bool           `koanf:"verbose" default:"false"`
	nets    map[int]string `koanf:"subnet"`
	Subnets []string
}

type RedisConfig struct {
	Address            string        `koanf:"address"`
	Username           string        `koanf:"username"`
	Password           string        `koanf:"password"`
	Database           int           `koanf:"database" default:"0"`
	ConnectionAttempts int           `koanf:"connectionAttempts" default:"3"`
	ConnectionCooldown time.Duration `koanf:"connectionCooldown" default:"1s"`
	TTL                time.Duration `koanf:"ttl" default:"20m"`
}

const prefix = "ARC_"

func Get() (*Config, error) {
	var res Config
	err := Load(prefix, &res)
	if err == nil {
		if len(res.Redis.Address) == 0 {
			err = fmt.Errorf("ARC_REDIS_ADDRESS has to be set")
		} else {
			if len(res.nets) > 0 {
				sub := make([]string, len(res.nets))
				for _, s := range res.nets {
					sub = append(sub, s)
				}
				res.Subnets = sub
				return &res, nil
			} else {
				err = fmt.Errorf("No subnet set")
			}
		}
	}
	return nil, err
}
