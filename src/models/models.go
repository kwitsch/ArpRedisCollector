package models

import (
	"net"

	"github.com/irai/arp"
)

// CacheMessage stores a MACEntry and additional flags
// Static: if enabled redis entries have no TTL
type CacheMessage struct {
	Entry  *arp.MACEntry
	Static bool
}

// IfNetPack bundles interface and network information
type IfNetPack struct {
	Interface *net.Interface
	Network   *net.IPNet
	Gateway   *net.IP
}
