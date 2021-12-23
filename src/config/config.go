package config

import (
	"fmt"
	"net"
	"time"

	. "github.com/kwitsch/go-dockerutils/config"
)

type Config struct {
	Redis   RedisConfig    `koanf:"redis"`
	Verbose bool           `koanf:"verbose" default:"false"`
	nets    map[int]string `koanf:"subnet"`
	Subnets []*net.IPMask
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
				smasks := make([]*net.IPMask, 0)

				var snet *net.IPNet

				for _, f := range res.nets {
					_, snet, err = net.ParseCIDR(f)
					if err == nil {
						smasks = append(smasks, &snet.Mask)
					} else {
						return nil, err
					}
				}

				res.Subnets = smasks

				return &res, nil
			} else {
				err = fmt.Errorf("No subnet set")
			}
		}
	}
	return nil, err
}
