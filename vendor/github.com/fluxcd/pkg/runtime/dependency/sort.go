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

package dependency

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"

	"github.com/fluxcd/pkg/runtime/internal/tarjan"
)

// Dependent provides an interface for resources that maintain
// CrossNamespaceDependencyReference list.
type Dependent interface {
	// GetDependsOn returns the Dependent's types.NamespacedName,
	// and the CrossNamespaceDependencyReference slice it depends on.
	GetDependsOn() (types.NamespacedName, []CrossNamespaceDependencyReference)
}

// CrossNamespaceDependencyReference holds the reference to a dependency.
type CrossNamespaceDependencyReference struct {
	// Namespace holds the namespace reference of a dependency.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Name holds the name reference of a dependency.
	// +required
	Name string `json:"name"`
}

func (r CrossNamespaceDependencyReference) String() string {
	if r.Namespace == "" {
		return r.Name
	}
	return fmt.Sprintf("%s%c%s", r.Namespace, types.Separator, r.Name)
}

// CircularDependencyError contains the circular dependency chains
// that were detected while sorting the Dependent dependencies.
type CircularDependencyError [][]string

func (e CircularDependencyError) Error() string {
	return fmt.Sprintf("circular dependencies: %v", [][]string(e))
}

// Sort sorts the Dependent slice based on their listed
// dependencies using Tarjan's strongly connected components algorithm.
func Sort(d []Dependent) ([]CrossNamespaceDependencyReference, error) {
	g, l := buildGraph(d)
	sccs := tarjan.SCC(g)
	var sorted []CrossNamespaceDependencyReference
	var circular CircularDependencyError
	for i := 0; i < len(sccs); i++ {
		s := sccs[i]
		if len(s) != 1 {
			circular = append(circular, s)
			continue
		}
		if n, ok := l[s[0]]; ok {
			sorted = append(sorted, n)
		}
	}
	if circular != nil {
		for i, j := 0, len(circular)-1; i < j; i, j = i+1, j-1 {
			circular[i], circular[j] = circular[j], circular[i]
		}
		return nil, circular
	}
	return sorted, nil
}

func buildGraph(d []Dependent) (tarjan.Graph, map[string]CrossNamespaceDependencyReference) {
	g := make(tarjan.Graph)
	l := make(map[string]CrossNamespaceDependencyReference)
	for i := 0; i < len(d); i++ {
		name, deps := d[i].GetDependsOn()
		g[name.String()] = buildEdges(deps, name.Namespace)
		l[name.String()] = CrossNamespaceDependencyReference(name)
	}
	return g, l
}

func buildEdges(d []CrossNamespaceDependencyReference, defaultNamespace string) tarjan.Edges {
	if len(d) == 0 {
		return nil
	}
	e := make(tarjan.Edges)
	for _, v := range d {
		if v.Namespace == "" {
			v.Namespace = defaultNamespace
		}
		e[v.String()] = struct{}{}
	}
	return e
}
