package changelog

const (
	Add    string = "+"
	Remove string = "-"
)

type DnsChange struct {
	Action     string
	Zone       string
	Record     string
	RecordType string
	Value      string
}

type ChangeLogger interface {
	LogChange(provider string, change DnsChange) error
}
