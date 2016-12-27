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

type ChangeLogger interface {
	LogChange(change DnsChange) error
}
