package collector

import (
	"context"
	"time"

	"github.com/irai/arp"
	"github.com/kwitsch/ArpRedisCollector/config"
	"github.com/kwitsch/ArpRedisCollector/models"
	arcnet "github.com/kwitsch/ArpRedisCollector/net"
)

const (
	probeInterval           time.Duration = time.Duration(1 * time.Minute)
	fullNetworkScanInterval time.Duration = time.Duration(10 * time.Minute)
	offlineDeadline         time.Duration = time.Duration(5 * time.Minute)
)

type Collector struct {
	cfg         *config.ArpConfig
	ctx         context.Context
	cancel      context.CancelFunc
	nethandlers []*NetHandler
	intChan     chan arp.MACEntry
	ArpChannel  chan *models.CacheMessage
}

type NetHandler struct {
	handler *arp.Handler
	ifNet   *models.IfNetPack
}

func New(cfg *config.ArpConfig) (*Collector, error) {
	if cfg.Verbose {
		arp.Debug = true
	}

	nets, err := arcnet.GetFilteredLocalNets(cfg.Subnets)
	if err == nil {
		var handlers []*NetHandler
		handlers, err = getAllHandlers(nets)
		if err == nil {
			ctx, cancel := context.WithCancel(context.Background())

			res := &Collector{
				cfg:         cfg,
				ctx:         ctx,
				cancel:      cancel,
				nethandlers: handlers,
				intChan:     make(chan arp.MACEntry, 256),
				ArpChannel:  make(chan *models.CacheMessage, 256),
			}

			return res, nil
		}
	}
	return nil, err
}

func (c *Collector) Close() {

}

func (c *Collector) start() {
	for _, h := range c.nethandlers {
		h.handler.AddNotificationChannel(c.intChan)
		go h.handler.ListenAndServe(c.ctx)
		c.ArpChannel <- c.getSelfCacheMessage(h)
	}
}

func (c *Collector) getSelfCacheMessage(h *NetHandler) *models.CacheMessage {
	res := &models.CacheMessage{
		Entry: &arp.MACEntry{
			MAC: h.ifNet.Interface.HardwareAddr,
			IPArray: [4]arp.IPEntry{
				{
					IP:          h.ifNet.Network.IP.To4(),
					LastUpdated: time.Now(),
				},
			},
		},
		Static: true,
	}

	return res
}

func getAllHandlers(nps []*models.IfNetPack) ([]*NetHandler, error) {
	res := make([]*NetHandler, 0)
	for _, np := range nps {
		h, err := getHandler(np)
		if err != nil {
			return nil, err
		}

		res = append(res, h)
	}
	return res, nil
}

func getHandler(np *models.IfNetPack) (*NetHandler, error) {
	cfg := arp.Config{
		NIC:                     np.Interface.Name,
		HostMAC:                 np.Interface.HardwareAddr,
		HostIP:                  np.Network.IP.To4(),
		RouterIP:                *np.Gateway,
		HomeLAN:                 *np.Network,
		ProbeInterval:           probeInterval,
		FullNetworkScanInterval: fullNetworkScanInterval,
		OfflineDeadline:         offlineDeadline,
	}

	handler, err := arp.New(cfg)
	if err == nil {
		res := &NetHandler{
			handler: handler,
			ifNet:   np,
		}

		return res, nil
	}

	return nil, err
}

/*
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
}*/
