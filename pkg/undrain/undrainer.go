package undrain

import "github.com/czerwonk/dns-drain/pkg/changelog"

type Undrainer interface {
	Undrain(changes *changelog.DnsChangeSet) error
}
