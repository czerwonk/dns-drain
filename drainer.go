package main

import "net"

type Drainer interface {
	drain(ipNet *net.IPNet) error
}
