package collector

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/j-keck/arping"
	"github.com/kwitsch/ArpRedisCollector/config"
	"github.com/kwitsch/ArpRedisCollector/models"
	arcnet "github.com/kwitsch/ArpRedisCollector/net"
)

type Collector struct {
	cfg          *config.ArpConfig
	ctx          context.Context
	cancel       context.CancelFunc
	netpacks     []*models.IfNetPack
	pollInterval *time.Duration
	reqChannel   chan *net.IP
	ArpChannel   chan *models.CacheMessage
}

func New(cfg *config.ArpConfig) (*Collector, error) {
	if cfg.Verbose {
		arping.EnableVerboseLog()
	}
	nets, err := arcnet.GetFilteredLocalNets(cfg.Subnets)
	if err == nil {
		arping.SetTimeout(cfg.Timeout)

		ctx, cancel := context.WithCancel(context.Background())

		ips := 0
		for _, n := range nets {
			ips += len(n.Others)
		}
		pi := time.Duration(ips) * cfg.Timeout
		pi += cfg.Cooldown

		res := &Collector{
			cfg:          cfg,
			ctx:          ctx,
			cancel:       cancel,
			netpacks:     nets,
			pollInterval: &pi,
			reqChannel:   make(chan *net.IP, 10000),
			ArpChannel:   make(chan *models.CacheMessage, 256),
		}

		return res, nil
	}

	return nil, err
}

func (c *Collector) Close() {
	close(c.reqChannel)
	close(c.ArpChannel)
	c.cancel()
}

func (c *Collector) Start() {
	if c.cfg.Verbose {
		fmt.Println("Collector Start for:")
		for _, p := range c.netpacks {
			fmt.Println("-", p.String())
		}
	}
	c.poll()

	go func() {
		pollTicker := time.NewTicker(*c.pollInterval).C
		for {
			select {
			case rr := <-c.reqChannel:
				c.resolve(rr)
			case <-pollTicker:
				c.poll()
			case <-c.ctx.Done():
				return
			}
		}
	}()
}

func (c *Collector) poll() {
	if c.cfg.Verbose {
		fmt.Println("Collector poll")
	}

	for _, p := range c.netpacks {
		for _, ip := range p.Others {
			c.reqChannel <- ip
		}
	}
}

func (c *Collector) resolve(ip *net.IP) {
	addr, _, err := arping.Ping(*ip)
	if err == nil {
		c.ArpChannel <- &models.CacheMessage{
			IP:  ip,
			Mac: addr,
		}
		if c.cfg.Verbose {
			fmt.Println("Collector poll collected", ip.String(), "=", addr.String())
		}
	} else {
		if c.cfg.Verbose {
			fmt.Println("Collector poll error", err)
		}
	}
}
