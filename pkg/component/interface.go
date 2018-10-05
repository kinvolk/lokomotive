package component

type Interface interface {
	Name() string
	Install() error
}
