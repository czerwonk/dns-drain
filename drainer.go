package main

import "net"

type Drainer interface {
	DrainWithIpNet(ipNet *net.IPNet, newIp net.IP) error
	DrainWithValue(value string, newValue string) error
}
