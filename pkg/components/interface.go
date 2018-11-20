package components

type ComponentChanger interface {
	GetValues([]byte) (string, error)
}
