/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"path"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Artifact represents the output of a source synchronisation.
type Artifact struct {
	// Path is the relative file path of this artifact.
	// +required
	Path string `json:"path"`

	// URL is the HTTP address of this artifact.
	// +required
	URL string `json:"url"`

	// Revision is a human readable identifier traceable in the origin source
	// system. It can be a Git commit SHA, Git tag, a Helm index timestamp, a Helm
	// chart version, etc.
	// +optional
	Revision string `json:"revision"`

	// Checksum is the SHA1 checksum of the artifact.
	// +optional
	Checksum string `json:"checksum"`

	// LastUpdateTime is the timestamp corresponding to the last update of this
	// artifact.
	// +required
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

// HasRevision returns true if the given revision matches the current Revision
// of the Artifact.
func (in *Artifact) HasRevision(revision string) bool {
	if in == nil {
		return false
	}
	return in.Revision == revision
}

// ArtifactDir returns the artifact dir path in the form of
// <source-kind>/<source-namespace>/<source-name>.
func ArtifactDir(kind, namespace, name string) string {
	kind = strings.ToLower(kind)
	return path.Join(kind, namespace, name)
}

// ArtifactPath returns the artifact path in the form of
// <source-kind>/<source-namespace>/<source-name>/<artifact-filename>.
func ArtifactPath(kind, namespace, name, filename string) string {
	return path.Join(ArtifactDir(kind, namespace, name), filename)
}
