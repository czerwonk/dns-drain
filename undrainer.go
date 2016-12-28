package main

import "github.com/czerwonk/dns-drain/changelog"

type Undrainer interface {
	Undrain(changes *changelog.DnsChangeSet) error
}
