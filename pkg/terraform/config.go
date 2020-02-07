package terraform

// A configuration struct for program information
// such as the current working directory and
// command-line arguments.
type Config struct {
	WorkingDir string
	Quiet      bool
}
