// SPDX-FileCopyrightText: (c) 2016 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package drain

import (
	"net"
	"regexp"
)

type Drainer interface {
	DrainWithIpNet(ipNet *net.IPNet, newIp net.IP) error
	DrainWithValue(value string, newValue string) error
	DrainWithRegex(regex *regexp.Regexp, newValue string) error
}
