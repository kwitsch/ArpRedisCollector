package collector

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/kwitsch/ArpRedisCollector/config"
	"github.com/kwitsch/ArpRedisCollector/models"
	arcnet "github.com/kwitsch/ArpRedisCollector/net"
	"github.com/mdlayher/arp"
)

type Collector struct {
	cfg         *config.ArpConfig
	ctx         context.Context
	cancel      context.CancelFunc
	nethandlers []*NetHandler
	polldur     time.Duration
	reqChannel  chan *resolveRequest
	ArpChannel  chan *models.CacheMessage
}

type NetHandler struct {
	client *arp.Client
	ifNet  *models.IfNetPack
}

type resolveRequest struct {
	client *arp.Client
	ip     *net.IP
}

func New(cfg *config.ArpConfig) (*Collector, error) {
	nets, err := arcnet.GetFilteredLocalNets(cfg.Subnets)
	if err == nil {
		var handlers []*NetHandler
		handlers, err = getAllHandlers(nets, cfg)
		if err == nil {
			ctx, cancel := context.WithCancel(context.Background())

			ic := 0
			for _, h := range handlers {
				ic += len(h.ifNet.Others)
			}
			if cfg.Verbose {
				fmt.Println("Collector created for", ic, "addresses")
			}

			res := &Collector{
				cfg:         cfg,
				ctx:         ctx,
				cancel:      cancel,
				nethandlers: handlers,
				polldur:     (time.Duration(ic) * cfg.Timeout) + cfg.Intervall,
				reqChannel:  make(chan *resolveRequest, 10000),
				ArpChannel:  make(chan *models.CacheMessage, 256),
			}

			return res, nil
		}
	}
	return nil, err
}

func (c *Collector) Close() {
	for _, h := range c.nethandlers {
		h.client.Close()
	}
	close(c.reqChannel)
	close(c.ArpChannel)
	c.cancel()
}

func (c *Collector) Start() {
	if c.cfg.Verbose {
		fmt.Println("Collector Start for:")
		for _, h := range c.nethandlers {
			fmt.Println("-", h.ifNet.String())
		}
		fmt.Println("Polltimer:", c.polldur.String())
	} else {
		fmt.Println("Collector Start")
	}

	go c.poll()

	go func() {
		pollTicker := time.NewTicker(c.polldur).C
		for {
			select {
			case rr := <-c.reqChannel:
				c.resolve(rr)
			case <-pollTicker:
				c.poll()
			case <-c.ctx.Done():
				fmt.Println("Collector Close")
				return
			}
		}
	}()
}

func (c *Collector) poll() {
	fmt.Println("Collector poll")

	for _, h := range c.nethandlers {
		c.handlerPoll(h)
	}
}

func (c *Collector) handlerPoll(h *NetHandler) {
	c.setSelf(h)
	for _, ip := range h.ifNet.Others {
		c.reqChannel <- &resolveRequest{
			ip:     ip,
			client: h.client,
		}
	}
}

func (c *Collector) setSelf(h *NetHandler) {
	c.publish(h.ifNet.IP, h.client.HardwareAddr())
}

func (c *Collector) resolve(rr *resolveRequest) {
	if c.cfg.Verbose {
		fmt.Println("Resolve", rr.ip.String())
	}
	rr.client.SetDeadline(time.Now().Add(c.cfg.Timeout))
	addr, err := rr.client.Resolve(*rr.ip)
	if err == nil {
		c.publish(rr.ip, addr)
		fmt.Println(rr.ip.String(), "=", addr.String())
	} else {
		if c.cfg.Verbose {
			fmt.Println(err)
		}
	}
}

func (c *Collector) publish(ip *net.IP, mac net.HardwareAddr) {
	c.ArpChannel <- &models.CacheMessage{
		IP:     ip,
		Mac:    mac,
		Static: c.cfg.StaticTable,
	}
}

func getAllHandlers(nps []*models.IfNetPack, cfg *config.ArpConfig) ([]*NetHandler, error) {
	res := make([]*NetHandler, 0)
	for _, np := range nps {
		h, err := getHandler(np, cfg)
		if err != nil {
			return nil, err
		}

		res = append(res, h)
	}
	return res, nil
}

func getHandler(np *models.IfNetPack, cfg *config.ArpConfig) (*NetHandler, error) {
	c, err := arp.Dial(&np.Interface)
	if err == nil {
		res := &NetHandler{
			client: c,
			ifNet:  np,
		}

		return res, nil
	}

	return nil, err
}
