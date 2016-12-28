package changelog

type ChangeLogger interface {
	LogChange(change DnsChange) error
}
