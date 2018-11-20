package ingress

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/utils"
)

const name = "nginx-ingress"

func init() {
	components.Register(name, &Answers{})
}

// Answers struct defines what all values can be provided to this component to
// tweak in it's answers file
type Answers struct {
	Namespace string `json:"namespace"`
}

// GetValues takes in answers file as array of bytes and returns the renderd
// value as string, otherwise returns an error
func (a *Answers) GetValues(data []byte) (string, error) {
	a, err := readAnswers(data)
	if err != nil {
		return "", errors.Wrap(err, "could not read the answers file")
	}
	return utils.RenderTemplate(values, a)
}

// readAnswers takes in answer's file contents as byte array and returns parsed
// Answers object
func readAnswers(data []byte) (*Answers, error) {
	var a Answers
	if err := yaml.Unmarshal(data, &a); err != nil {
		return nil, err
	}
	return &a, nil
}
