package models

import (
	"fmt"
	"net"
)

// CacheMessage stores a MACEntry and additional flags
// Static: if enabled redis entries have no TTL
type CacheMessage struct {
	Mac    net.HardwareAddr
	IP     *net.IP
	Static bool
}

// IfNetPack bundles interface and network information
type IfNetPack struct {
	Interface *net.Interface
	Network   *net.IPNet
	IP        *net.IP
	Others    []*net.IP
}

// String returns a string representation of the stack
// Format: "Interface: %s(%s) | Network: %s | IP: %s | Others: %s,...""
func (pack IfNetPack) String() string {
	return fmt.Sprintf("Interface: %s(%s) | Network: %s | IP: %s | Others: %s,...",
		pack.Interface.Name,
		pack.Interface.HardwareAddr.String(),
		pack.Network.String(),
		pack.IP.String(),
		pack.Others[0].String())
}
