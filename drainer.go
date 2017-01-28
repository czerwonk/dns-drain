package main

import "net"

type Drainer interface {
	Drain(ipNet *net.IPNet, newIp net.IP) error
}
