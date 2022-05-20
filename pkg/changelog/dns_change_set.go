package changelog

const (
	Add    string = "+"
	Remove string = "-"
)

type DnsChangeSet struct {
	Changes []DnsChange `json:"changes"`
}

type DnsChange struct {
	Provider   string `json:"provider"`
	Action     string `json:"action"`
	Zone       string `json:"zone"`
	Record     string `json:"record"`
	RecordType string `json:"recordType"`
	Value      string `json:"value"`
}

func (c *DnsChangeSet) GroupByZone() map[string][]DnsChange {
	m := make(map[string][]DnsChange)

	for _, change := range c.Changes {
		var arr []DnsChange
		var found bool
		if arr, found = m[change.Zone]; !found {
			arr = make([]DnsChange, 0)
		}

		m[change.Zone] = append(arr, change)
	}

	return m
}
