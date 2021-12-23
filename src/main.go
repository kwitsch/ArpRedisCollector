package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/kwitsch/ArpRedisCollector/collector"
	"github.com/kwitsch/ArpRedisCollector/config"
	"github.com/kwitsch/ArpRedisCollector/redis"

	_ "github.com/kwitsch/go-dockerutils"
)

func main() {
	cfg, cErr := config.Get()
	if cErr == nil {
		redis, rErr := redis.New(&cfg.Redis)
		if rErr == nil {
			arp, aErr := collector.New(&cfg.Arp)
			if aErr == nil {
				fmt.Println("Collector start")

				intChan := make(chan os.Signal, 1)
				signal.Notify(intChan, os.Interrupt)

				ticker := time.NewTicker(cfg.Arp.FullNetworkScanInterval).C
				for {
					select {
					case a := <-arp.ArpChannel:
						redis.Publish(&a)
					case <-ticker:
						arp.PublishTable()
					case <-intChan:
						fmt.Println("Collector stopping")
						arp.Close()
						redis.Close()
						os.Exit(0)
					}
				}
			} else {
				fmt.Println(aErr)
				os.Exit(3)
			}
		} else {
			fmt.Println(rErr)
			os.Exit(2)
		}
	} else {
		fmt.Println(cErr)
		os.Exit(1)
	}
}
