package certmanager

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/utils"
)

const name = "cert-manager"

func init() {
	components.Register(name, &Answers{})
}

// Answers struct defines what all values can be provided to this component to
// tweak in it's answers file
type Answers struct {
	Namespace string `json:"namespace"`
	Email     string `json:"email"`
}

// GetValues takes in answers file as array of bytes and returns the renderd
// value as string, otherwise returns an error
func (a *Answers) GetValues(data []byte) (string, error) {
	if err := yaml.Unmarshal(data, a); err != nil {
		return "", errors.Wrap(err, "could not read the answers file")
	}
	if a.Email == "" {
		return "", errors.New("missing mandatory answer email required for ACME registration")
	}

	return utils.RenderTemplate(values, a)
}
