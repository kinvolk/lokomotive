package util

import "encoding/json"

type NodeSelector struct {
	Key      string   `hcl:"key",json:"key,omitempty"`
	Operator string   `hcl:"operator",json:"operator,omitempty"`
	Values   []string `hcl:"values,optional",json:"values,omitempty"`
}

type Toleration struct {
	Key               string `hcl:"key,optional" json:"key,omitempty"`
	Effect            string `hcl:"effect,optional" json:"effect,omitempty"`
	Operator          string `hcl:"operator,optional" json:"operator,omitempty"`
	Value             string `hcl:"value,optional" json:"value,omitempty"`
	TolerationSeconds string `hcl:"toleration_seconds,optional" json:"toleration_seconds,omitempty"`
}

// RenderTolerations takes a list of tolerations.
// It returns a json string and an error if any.
func RenderTolerations(t []Toleration) (string, error) {
	if len(t) == 0 {
		return "", nil
	}

	b, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
