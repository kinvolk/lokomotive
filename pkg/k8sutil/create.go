// Copyright 2020 The Lokomotive Authors
// Copyright 2019 The Kubernetes Authors
// Copyright 2015 CoreOS, Inc
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

package k8sutil

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

// Adapted from https://github.com/kubernetes-incubator/bootkube/blob/83d32756c6b02c26cab1de3f03b57f06ae4339a7/pkg/bootkube/create.go

type manifest struct {
	kind       string
	apiVersion string
	namespace  string
	name       string
	raw        []byte

	filepath string
}

func (m manifest) String() string {
	if m.namespace == "" {
		return fmt.Sprintf("%s %s %s", m.filepath, m.kind, m.name)
	}
	return fmt.Sprintf("%s %s %s/%s", m.filepath, m.kind, m.namespace, m.name)
}

func (m manifest) Kind() string {
	return m.kind
}

func (m manifest) Raw() []byte {
	return m.raw
}

// LoadManifests parses a map of Kubernetes manifests.
func LoadManifests(files map[string]string) ([]manifest, error) {
	var manifests []manifest
	for path, fileContent := range files {
		r := strings.NewReader(fileContent)
		ms, err := parseManifests(r)
		if err != nil {
			return nil, errors.Wrapf(err, "error parsing file %s:", path)
		}
		manifests = append(manifests, ms...)
	}
	return manifests, nil
}

// parseManifests parses a YAML or JSON document that may contain one or more
// kubernetes resources.
func parseManifests(r io.Reader) ([]manifest, error) {
	reader := yaml.NewYAMLReader(bufio.NewReader(r))
	var manifests []manifest
	for {
		yamlManifest, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				return manifests, nil
			}
			return nil, err
		}
		yamlManifest = bytes.TrimSpace(yamlManifest)
		if len(yamlManifest) == 0 {
			continue
		}

		m, err := parseYAMLManifest(yamlManifest)
		if err != nil {
			return nil, fmt.Errorf("parse manifest: %w", err)
		}
		manifests = append(manifests, m...)
	}
}

// parseYAMLManifest parses a single YAML Kubernetes resource.
func parseYAMLManifest(data []byte) ([]manifest, error) {
	if string(data) == "null" {
		return nil, nil
	}
	var m struct {
		APIVersion string `yaml:"apiVersion,omitempty"`
		Kind       string `yaml:"kind,omitempty"`
		Metadata   struct {
			Name      string `yaml:"name,omitempty"`
			Namespace string `yaml:"namespace,omitempty"`
		} `yaml:"metadata,omitempty"`
	}

	if err := k8syaml.Unmarshal(data, &m); err != nil {
		return nil, errors.Wrapf(err, "failed to parse manifest")
	}

	// We continue if the object we received was a *List kind. Otherwise if a
	// single object is received we just return from here.
	if !strings.HasSuffix(m.Kind, "List") {
		return []manifest{{
			kind:       m.Kind,
			apiVersion: m.APIVersion,
			namespace:  m.Metadata.Namespace,
			name:       m.Metadata.Name,
			raw:        data,
		}}, nil
	}

	// We parse the list of items and extract one object at a time
	var mList struct {
		APIVersion string `yaml:"apiVersion,omitempty"`
		Kind       string `yaml:"kind,omitempty"`
		Metadata   struct {
			Name      string `yaml:"name,omitempty"`
			Namespace string `yaml:"namespace,omitempty"`
		} `yaml:"metadata,omitempty"`
		Items []json.RawMessage `yaml:"items"`
	}

	if err := k8syaml.Unmarshal(data, &mList); err != nil {
		return nil, errors.Wrapf(err, "failed to parse manifest list")
	}

	var manifests []manifest
	for _, item := range mList.Items {
		// make a recursive call, since this is a single object it will be
		// parsed and returned to us
		mn, err := parseYAMLManifest(item)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, mn...)
	}
	return manifests, nil
}
