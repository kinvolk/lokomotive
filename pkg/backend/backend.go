package backend

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// Backend describes the Terraform state storage location.
type Backend interface {
	// LoadConfig loads the backend config provided by the user.
	LoadConfig(*hcl.Body, *hcl.EvalContext) hcl.Diagnostics
	// Render renders the backend template with user backend configuration.
	Render() (string, error)
	// Validate validates backend configuration.
	Validate() error
}

// backends is a collection in which all Backends get automatically registered.
var backends map[string]Backend

// Initialize package's global variable on import.
func init() {
	backends = make(map[string]Backend)
}

// Register registers Backend b in the internal backends map.
func Register(name string, b Backend) {
	if _, exists := backends[name]; exists {
		panic(fmt.Sprintf("backend with name %q registered already", name))
	}
	backends[name] = b
}

// GetBackend returns the Backend referred to by name.
func GetBackend(name string) (Backend, error) {
	backend, exists := backends[name]
	if !exists {
		return nil, fmt.Errorf("no backend with name %q found", name)
	}
	return backend, nil
}
