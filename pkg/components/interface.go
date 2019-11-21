package components

import (
	"github.com/hashicorp/hcl2/hcl"
)

// Component represents functionality each Lokomotive component should implement.
type Component interface {
	// LoadConfig loads specific component configuration from HCL configuration.
	// If an error occurs, HCL diagnostics are returned.
	LoadConfig(*hcl.Body, *hcl.EvalContext) hcl.Diagnostics
	// RenderManifests returns a map of Kubernetes manifests in YAML format, where
	// the key is the file from which the content comes.
	RenderManifests() (map[string]string, error)
	// Metadata returns component metadata.
	Metadata() Metadata
}
