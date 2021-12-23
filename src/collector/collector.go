package collector

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/irai/arp"
	"github.com/kwitsch/ArpRedisCollector/config"
	arcnet "github.com/kwitsch/ArpRedisCollector/net"
)

const (
	probeInterval           time.Duration = time.Duration(1 * time.Minute)
	fullNetworkScanInterval time.Duration = time.Duration(10 * time.Minute)
	offlineDeadline         time.Duration = time.Duration(5 * time.Minute)
)

type Collector struct {
	cfg        *config.ArpConfig
	verbose    bool
	ctx        context.Context
	cancel     context.CancelFunc
	handler    *arp.Handler
	network    *net.IPNet
	ArpChannel chan arp.MACEntry
}

func New(cfg *config.ArpConfig, verbose bool) (*Collector, error) {
	acfg, err := getConfig(cfg)
	if err == nil {
		if verbose {
			arp.Debug = true
		}

		var handler *arp.Handler
		handler, err = arp.New(*acfg)
		if err == nil {
			ctx, cancel := context.WithCancel(context.Background())
			arpChannel := make(chan arp.MACEntry, 256)
			res := &Collector{
				cfg:        cfg,
				verbose:    verbose,
				ctx:        ctx,
				cancel:     cancel,
				handler:    handler,
				network:    &acfg.HomeLAN,
				ArpChannel: arpChannel,
			}

			go res.handler.ListenAndServe(res.ctx)

			res.handler.AddNotificationChannel(res.ArpChannel)

			return res, nil
		}
	}
	return nil, err
}

func (c *Collector) Close() {
	c.cancel()

	c.handler.Close()

	close(c.ArpChannel)
}

func (c *Collector) PublishTable() {
	if c.verbose {
		fmt.Println("Collector.PublishTable")
		c.handler.PrintTable()
	}

	for _, entry := range c.handler.GetTable() {
		c.ArpChannel <- entry
	}
}

func getConfig(cfg *config.ArpConfig) (*arp.Config, error) {
	iface, err := net.InterfaceByName(cfg.Interface)
	if err == nil {
		homeNet := arcnet.GetHomeNet(iface)
		if homeNet != nil {
			var gateway net.IP
			gateway, err = arcnet.GetDefaultGateway(cfg.Interface)
			if err == nil {
				res := &arp.Config{
					NIC:                     iface.Name,
					HostMAC:                 iface.HardwareAddr,
					HostIP:                  homeNet.IP.To4(),
					RouterIP:                gateway,
					HomeLAN:                 *homeNet,
					ProbeInterval:           probeInterval,
					FullNetworkScanInterval: fullNetworkScanInterval,
					OfflineDeadline:         offlineDeadline,
				}
				return res, nil
			}
		} else {
			err = fmt.Errorf("%s has no valid IPv4 address", cfg.Interface)
		}
	}
	return nil, err
}
