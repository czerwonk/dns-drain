package changelog

const (
	Add    string = "+"
	Remove string = "-"
)

type DnsChange struct {
	Provider   string
	Action     string
	Zone       string
	Record     string
	RecordType string
	Value      string
}

type ChangeLogger interface {
	LogChange(change DnsChange) error
}
