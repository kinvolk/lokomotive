package cluster

// Manager interface provide methods that  manage the lifecycle
// of a Lokomotive cluster
type Manager interface {
	Apply(*Options) error
	Destroy(*Options) error
	ApplyComponents([]string) error
	RenderComponents([]string) error
	Health() error
}

// Options struct represents the CLI options
type Options struct {
	Confirm         bool
	UpgradeKubelets bool
	SkipComponents  bool
	Verbose         bool
}
