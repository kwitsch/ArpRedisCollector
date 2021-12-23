package models

import "github.com/irai/arp"

// CacheMessage stores a MACEntry and additional flags
// Static: if enabled redis entries have no TTL
type CacheMessage struct {
	Entry  *arp.MACEntry
	Static bool
}
