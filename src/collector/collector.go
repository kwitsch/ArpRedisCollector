package collector

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/kwitsch/ArpRedisCollector/config"
	"github.com/kwitsch/ArpRedisCollector/models"
	arcnet "github.com/kwitsch/ArpRedisCollector/net"
	marp "github.com/mdlayher/arp"
)

type Collector struct {
	cfg         *config.ArpConfig
	ctx         context.Context
	cancel      context.CancelFunc
	nethandlers []*NetHandler
	ArpChannel  chan *models.CacheMessage
}

type NetHandler struct {
	client *marp.Client
	ifNet  *models.IfNetPack
}

func New(cfg *config.ArpConfig) (*Collector, error) {
	nets, err := arcnet.GetFilteredLocalNets(cfg.Subnets)
	if err == nil {
		var handlers []*NetHandler
		handlers, err = getAllHandlers(nets, cfg)
		if err == nil {
			ctx, cancel := context.WithCancel(context.Background())

			res := &Collector{
				cfg:         cfg,
				ctx:         ctx,
				cancel:      cancel,
				nethandlers: handlers,
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
	close(c.ArpChannel)
}

func (c *Collector) Start() {
	c.poll()

	go func() {
		pollTicker := time.NewTicker(c.cfg.PollIntervall).C
		readTicker := time.NewTicker(c.cfg.ReadIntervall).C
		for {
			select {
			case <-pollTicker:
				c.poll()
			case <-readTicker:
				c.read()
			}
		}
	}()
}

func (c *Collector) poll() {
	if c.cfg.Verbose {
		fmt.Println("Collector poll")
	}
	for _, h := range c.nethandlers {
		mask := binary.BigEndian.Uint32(h.ifNet.Network.Mask)
		start := binary.BigEndian.Uint32(h.ifNet.Network.IP)
		finish := (start & mask) | (mask ^ 0xffffffff)

		for i := start + 1; i < finish; i++ {
			ip := make(net.IP, 4)
			binary.BigEndian.PutUint32(ip, i)
			err := h.client.Request(ip)
			if err != nil {
				fmt.Println("Collector poll error", err, ip.String())
			}
		}
	}
}
func (c *Collector) read() {
	if c.cfg.Verbose {
		fmt.Println("Collector read")
	}
	for _, h := range c.nethandlers {
		h.client.SetDeadline(time.Now().Add(time.Second))
		arp, _, err := h.client.Read()
		if err == nil {
			if c.cfg.Verbose {
				fmt.Println("Collector read", arp.SenderHardwareAddr.String(), "=", arp.SenderIP.String())
			}
			c.ArpChannel <- &models.CacheMessage{
				IP:     arp.SenderIP,
				Mac:    arp.SenderHardwareAddr,
				Static: false,
			}
		} else {
			if c.cfg.Verbose {
				fmt.Println("Collector read error", err)
			}
		}
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
	c, err := marp.Dial(np.Interface)
	if err == nil {
		res := &NetHandler{
			client: c,
			ifNet:  np,
		}

		return res, nil
	}

	return nil, err
}
