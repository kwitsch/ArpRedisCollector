package config

import (
	"strings"
	"time"

	"github.com/creasty/defaults"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/env"
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

func Load() (*Config, error) {
	var result Config
	err := defaults.Set(&result)
	if err == nil {
		var k = koanf.New(".")
		k.Load(env.Provider(prefix, ".", func(s string) string {
			return strings.Replace(strings.ToLower(
				strings.TrimPrefix(s, prefix)), "_", ".", -1)
		}), nil)
		err = k.UnmarshalWithConf("", &result, koanf.UnmarshalConf{Tag: "koanf"})
		if err == nil {
			return &result, nil
		}
	}
	return nil, err
}
