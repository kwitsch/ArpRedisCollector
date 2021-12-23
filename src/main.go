package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/kwitsch/ArpRedisCollector/collector"
	"github.com/kwitsch/ArpRedisCollector/config"
	"github.com/kwitsch/ArpRedisCollector/redis"

	_ "github.com/kwitsch/go-dockerutils"
)

func main() {
	cfg, cErr := config.Get()
	if cErr == nil {
		redis, rErr := redis.New(&cfg.Redis, cfg.Arp.OfflineDeadline*2)
		if rErr == nil {
			arp, aErr := collector.New(&cfg.Arp)
			if aErr == nil {
				fmt.Println("Collector start")
				intChan := make(chan os.Signal, 1)
				signal.Notify(intChan, os.Interrupt)
				for {
					select {
					case a := <-arp.ArpChannel:
						redis.Publish(&a)
					case <-intChan:
						fmt.Println("Collector stopping")
						arp.Close()
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
