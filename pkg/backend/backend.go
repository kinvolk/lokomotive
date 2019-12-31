package backend

import (
	"fmt"

	"github.com/hashicorp/hcl2/hcl"
)

// Backend describes the storage place of terraform state.
type Backend interface {
	// LoadConfig loads the backend config provided by the user.
	LoadConfig(*hcl.Body, *hcl.EvalContext) hcl.Diagnostics
	//Render renders the backend template with user backend configuration
	Render() (string, error)
	// Validates backend configuration
	Validate() error
}

// backends is a collection where all backends gets automatically registered
var backends map[string]Backend

// initialize package's global variable when package is imported
func init() {
	backends = make(map[string]Backend)
}

// Register adds backend into internal map
func Register(name string, b Backend) {
	if _, exists := backends[name]; exists {
		panic(fmt.Sprintf("backend with name %q registered already", name))
	}
	backends[name] = b
}

// GetBackend returns backend based on the name
func GetBackend(name string) (Backend, error) {
	backend, exists := backends[name]
	if !exists {
		return nil, fmt.Errorf("no backend with name %q found", name)
	}
	return backend, nil
}
