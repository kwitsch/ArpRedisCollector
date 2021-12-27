package config

import (
	"fmt"
	"net"
	"time"

	. "github.com/kwitsch/go-dockerutils/config"
)

type Config struct {
	Redis   RedisConfig `koanf:"redis"`
	Arp     ArpConfig   `koanf:"arp"`
	Verbose bool        `koanf:"verbose" default:"false"`
}

type RedisConfig struct {
	Address            string        `koanf:"address"`
	Username           string        `koanf:"username"`
	Password           string        `koanf:"password"`
	Database           int           `koanf:"database" default:"0"`
	ConnectionAttempts int           `koanf:"connectionAttempts" default:"3"`
	ConnectionCooldown time.Duration `koanf:"connectionCooldown" default:"1s"`
	TTL                time.Duration
	Verbose            bool
}

type ArpConfig struct {
	CIDRs         map[int]string `koanf:"subnet"`
	PollIntervall time.Duration  `koan:"pollIntervall" default:"10m"`
	Readntervall  time.Duration  `koan:"readIntervall" default:"2s"`
	Subnets       []*net.IPNet
	Verbose       bool
}

const prefix = "ARC_"

func Get() (*Config, error) {
	var res Config
	err := Load(prefix, &res)
	if err == nil {
		if len(res.Redis.Address) == 0 {
			err = fmt.Errorf("ARC_REDIS_ADDRESS has to be set")
		} else {
			if len(res.Arp.CIDRs) > 0 {
				snets := make([]*net.IPNet, 0)

				var snet *net.IPNet

				for _, f := range res.Arp.CIDRs {
					_, snet, err = net.ParseCIDR(f)
					if err == nil {
						snets = append(snets, snet)
					} else {
						return nil, err
					}
				}

				res.Arp.Subnets = snets

				res.Redis.TTL = res.Arp.PollIntervall * 2

				res.Arp.Verbose = res.Verbose
				res.Redis.Verbose = res.Verbose

				return &res, nil
			} else {
				err = fmt.Errorf("No subnet set")
			}
		}
	}
	return nil, err
}
