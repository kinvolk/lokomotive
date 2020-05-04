// Copyright 2020 The Lokomotive Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import "encoding/json"

// NodeAffinity is a struct that other components can use to define the HCL format of NodeAffinity
// in Kubernetes PodSpec.
type NodeAffinity struct {
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
