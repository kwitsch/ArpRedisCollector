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
	cfg        *config.ArpConfig
	ctx        context.Context
	cancel     context.CancelFunc
	netpacks   []*models.IfNetPack
	reqChannel chan *resolveRequest
	ArpChannel chan *models.CacheMessage
}

type resolveRequest struct {
	intf *net.Interface
	ip   *net.IP
}

func New(cfg *config.ArpConfig) (*Collector, error) {
	if cfg.Verbose {
		arping.EnableVerboseLog()
	}
	nets, err := arcnet.GetFilteredLocalNets(cfg.Subnets)
	if err == nil {
		ctx, cancel := context.WithCancel(context.Background())

		res := &Collector{
			cfg:        cfg,
			ctx:        ctx,
			cancel:     cancel,
			netpacks:   nets,
			reqChannel: make(chan *resolveRequest, 1000),
			ArpChannel: make(chan *models.CacheMessage, 256),
		}

		return res, nil
	}

	return nil, err
}

func (c *Collector) Close() {
	close(c.reqChannel)
	close(c.ArpChannel)
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
		pollTicker := time.NewTicker(c.cfg.PollIntervall).C
		for {
			select {
			case rr := <-c.reqChannel:
				c.resolve(rr)
			case <-pollTicker:
				c.poll()
			}
		}
	}()
}

func (c *Collector) poll() {
	if c.cfg.Verbose {
		fmt.Println("Collector poll")
	}

	for _, p := range c.netpacks {
		c.setSelf(p)
		for _, ip := range p.Others {
			c.reqChannel <- &resolveRequest{
				ip:   ip,
				intf: p.Interface,
			}
		}
	}
}

func (c *Collector) setSelf(p *models.IfNetPack) {
	c.ArpChannel <- &models.CacheMessage{
		IP:     p.IP,
		Mac:    p.Interface.HardwareAddr,
		Static: true,
	}
}

func (c *Collector) resolve(rr *resolveRequest) {
	addr, _, err := arping.PingOverIface(*rr.ip, *rr.intf)
	if err == nil {
		c.ArpChannel <- &models.CacheMessage{
			IP:     rr.ip,
			Mac:    addr,
			Static: c.cfg.StaticTable,
		}
		if c.cfg.Verbose {
			fmt.Println("Collector poll collected", rr.ip.String(), "=", addr.String())
		}
	} else {
		if c.cfg.Verbose {
			fmt.Println("Collector poll error", err, rr.ip.String())
		}
	}
}
