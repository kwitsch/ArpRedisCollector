package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/kwitsch/ArpRedisCollector/collector"
	"github.com/kwitsch/ArpRedisCollector/config"
)

func main() {
	cfg, cErr := config.Get()
	if cErr == nil {
		arp, aErr := collector.New(&cfg.Arp)
		if aErr == nil {
			fmt.Println("Collector start")
			intChan := make(chan os.Signal, 1)
			signal.Notify(intChan, os.Interrupt)
			for {
				select {
				case a := <-arp.ArpChannel:
					fmt.Println(a.String())
				case <-intChan:
					fmt.Println("Collector stopping")
					arp.Close()
					os.Exit(0)
				}
			}
		} else {
			fmt.Println(aErr)
			os.Exit(2)
		}
	} else {
		fmt.Println(cErr)
		os.Exit(1)
	}
}
