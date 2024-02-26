package utils

import "net"

// CIDRContains checks if the given IP is in the CIDR
func CIDRContains(cidr []*net.IPNet, ip string) bool {
	for _, c := range cidr {
		if c.Contains(net.ParseIP(ip)) {
			return true
		}
	}
	return false
}
