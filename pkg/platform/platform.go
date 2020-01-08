package platform

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// Platform describes single environment, where cluster can be installed
type Platform interface {
	LoadConfig(*hcl.Body, *hcl.EvalContext) hcl.Diagnostics
	Install() error
	Destroy() error
	GetAssetDir() string
	GetExpectedNodes() int
}

// platforms is a collection where all platforms gets automatically registered
var platforms map[string]Platform

// initialize package's global variable when package is imported
func init() {
	platforms = make(map[string]Platform)
}

// Register adds platform into internal map
func Register(name string, p Platform) {
	if _, exists := platforms[name]; exists {
		panic(fmt.Sprintf("platform with name %q registered already", name))
	}
	platforms[name] = p
}

// GetPlatform returns platform based on the name
func GetPlatform(name string) (Platform, error) {
	platform, exists := platforms[name]
	if !exists {
		return nil, fmt.Errorf("no platform with name %q found", name)
	}
	return platform, nil
}
