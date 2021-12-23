package net

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

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

func GetDefaultGateway(ifName string) (gw net.IP, err error) {
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
