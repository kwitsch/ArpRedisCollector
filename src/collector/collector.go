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
	intChan     chan *intRequest
	ArpChannel  chan *models.CacheMessage
}

type NetHandler struct {
	client *marp.Client
	ifNet  *models.IfNetPack
}

type intRequest struct {
	client marp.Client
	ip     net.IP
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
				intChan:     make(chan *intRequest, 256),
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
	close(c.intChan)
	close(c.ArpChannel)
}

func (c *Collector) Start() {
	if c.cfg.Verbose {
		fmt.Println("Collector start")
	}
	c.poll()

	go func() {
		ticker := time.NewTicker(c.cfg.PollIntervall).C
		for {
			select {
			case <-ticker:
				if c.cfg.Verbose {
					fmt.Println("Collector poll")
				}
				c.poll()
			case m := <-c.intChan:
				ha, err := m.client.Resolve(m.ip)
				if err == nil {
					c.ArpChannel <- &models.CacheMessage{
						Mac:    ha,
						IP:     m.ip,
						Static: false,
					}
				}
			}
		}
	}()
}

func (c *Collector) poll() {
	for _, h := range c.nethandlers {
		mask := binary.BigEndian.Uint32(h.ifNet.Network.Mask)
		start := binary.BigEndian.Uint32(h.ifNet.Network.IP)
		finish := (start & mask) | (mask ^ 0xffffffff)

		for i := start + 1; i < finish; i++ {
			ip := make(net.IP, 4)
			binary.BigEndian.PutUint32(ip, i)
			c.intChan <- &intRequest{
				ip:     ip,
				client: *h.client,
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
