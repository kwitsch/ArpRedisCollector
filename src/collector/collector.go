package collector

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/irai/arp"
	"github.com/kwitsch/docker-ArpRedisCollector/config"
)

type Collector struct {
	cfg        *config.ArpConfig
	handler    *arp.Handler
	ctx        context.Context
	cancel     context.CancelFunc
	ArpChannel chan arp.MACEntry
}

func New(cfg *config.ArpConfig) (*Collector, error) {
	acfg, err := getConfig(cfg)
	if err == nil {
		var handler *arp.Handler
		handler, err = arp.New(*acfg)
		if err == nil {
			ctx, cancel := context.WithCancel(context.Background())
			arpChannel := make(chan arp.MACEntry, 16)
			res := &Collector{
				cfg:        cfg,
				handler:    handler,
				ctx:        ctx,
				cancel:     cancel,
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

func getConfig(cfg *config.ArpConfig) (*arp.Config, error) {
	iface, err := net.InterfaceByName(cfg.Interface)
	if err == nil {
		homeNet := getHomeNet(iface)
		if homeNet != nil {
			var gateway net.IP
			gateway, err = getDefaultGateway(cfg.Interface)
			if err == nil {
				return &arp.Config{
					NIC:                     iface.Name,
					HostMAC:                 iface.HardwareAddr,
					HostIP:                  homeNet.IP.To4(),
					RouterIP:                gateway,
					HomeLAN:                 *homeNet,
					ProbeInterval:           cfg.ProbeInterval,
					FullNetworkScanInterval: cfg.FullNetworkScanInterval,
					PurgeDeadline:           cfg.PurgeDeadline,
				}, nil
			}
		} else {
			err = fmt.Errorf("%s has no valid IPv4 address", cfg.Interface)
		}
	}
	return nil, err
}

func getHomeNet(iface *net.Interface) *net.IPNet {
	addrs, ierr := iface.Addrs()
	if ierr == nil {
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				if strings.Count(v.String(), ":") < 2 && !v.IP.IsLoopback() {
					return v
				}
			}
		}
	}
	return nil
}

func getDefaultGateway(ifName string) (gw net.IP, err error) {

	file, err := os.Open("/proc/net/route")
	if err != nil {
		return net.IPv4zero, err
	}
	defer file.Close()

	res := net.IPv4zero
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// 0 = Iface
		// 1 = Destination
		// 2 = Gateway
		fields := strings.Split(scanner.Text(), "\t")
		if fields[0] == ifName {
			d64, _ := strconv.ParseInt("0x"+fields[2], 0, 64)
			if d64 != 0 {
				d32 := uint32(d64)
				res = make(net.IP, 4)
				binary.LittleEndian.PutUint32(res, d32)
				break
			}
		}
	}
	if res.Equal(net.IPv4zero) {
		err = fmt.Errorf("Coulden't get default gateway for interface %s", ifName)
	}
	return res, err
}
