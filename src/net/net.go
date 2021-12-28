package net

import (
	"encoding/binary"
	"fmt"
	"net"

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
					Network: hnet,
					IP:      &hnet.IP,
					Others:  GetAllIps(hnet, &hnet.IP),
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
				if v.IP.To4() != nil &&
					!v.IP.IsLoopback() {
					return v
				}
			}
		}
	}
	return nil
}

func GetAllIps(hnet *net.IPNet, self *net.IP) []*net.IP {
	res := make([]*net.IP, 0)

	_, v4net, _ := net.ParseCIDR(hnet.String())
	mask := binary.BigEndian.Uint32(v4net.Mask)
	start := binary.BigEndian.Uint32(v4net.IP)
	finish := (start & mask) | (mask ^ 0xffffffff)

	for i := start + 1; i < finish; i++ {
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		res = append(res, &ip)
	}

	return res
}
