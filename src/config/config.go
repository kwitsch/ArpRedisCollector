package config

import (
	"fmt"
	"time"

	. "github.com/kwitsch/go-dockerutils/config"
)

type Config struct {
	Redis RedisConfig `koanf:"redis"`
	Arp   ArpConfig   `koanf:"arp"`
}

type RedisConfig struct {
	Address            string        `koanf:"address"`
	Username           string        `koanf:"username"`
	Password           string        `koanf:"password"`
	Database           int           `koanf:"database" default:"0"`
	ConnectionAttempts int           `koanf:"connectionAttempts" default:"3"`
	ConnectionCooldown time.Duration `koanf:"connectionCooldown" default:"1s"`
}

type ArpConfig struct {
	Interface               string        `koanf:"interface"`
	ProbeInterval           time.Duration `koanf:"probeInterval" default:"1m"`
	FullNetworkScanInterval time.Duration `koanf:"fullNetworkScanInterval" default:"20m"`
	PurgeDeadline           time.Duration `koanf:"purgeDeadline" default:"10m"`
}

const prefix = "ARC_"

func Get() (*Config, error) {
	var res Config
	err := Load(prefix, &res)
	if err == nil {
		if len(res.Redis.Address) == 0 {
			err = fmt.Errorf("REDIS_ADDRESS has to be set")
		} else if len(res.Arp.Interface) == 0 {
			err = fmt.Errorf("ARP_INTERFACE has to be set")
		} else {
			return &res, nil
		}
	}
	return nil, err
}
