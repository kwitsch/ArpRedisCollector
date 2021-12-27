package net

import (
	"fmt"
	"net"
	"strings"

	"github.com/kwitsch/ArpRedisCollector/models"
)

func GetAllLocalNets() ([]*models.IfNetPack, error) {
	res := make([]*models.IfNetPack, 0)
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, i := range ifaces {
			hnet := GetHomeNet(&i)
			if hnet != nil {
				aRes := &models.IfNetPack{
					Interface: &i,
					Network:   hnet,
				}
				res = append(res, aRes)
			}
		}
	}
	return res, err
}

func GetFilteredLocalNets(filters []*net.IPNet) ([]*models.IfNetPack, error) {
	res := make([]*models.IfNetPack, 0)
	nets, err := GetAllLocalNets()
	if err == nil {
		if len(nets) > 0 {
			for _, net := range nets {
				for _, f := range filters {
					if f.Contains(net.Network.IP) {
						res = append(res, net)
					}
				}
			}
		}
		if len(res) == 0 {
			err = fmt.Errorf("No matching local net found")
		}
	}
	return res, err
}

func GetHomeNet(iface *net.Interface) *net.IPNet {
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
