package models

import (
	"net"
)

// CacheMessage stores a MACEntry and additional flags
// Static: if enabled redis entries have no TTL
type CacheMessage struct {
	Mac    net.HardwareAddr
	IP     net.IP
	Static bool
}

// IfNetPack bundles interface and network information
type IfNetPack struct {
	Interface *net.Interface
	Network   *net.IPNet
}
