package backends

type Backend interface {
	Query()
	Exec()
}
