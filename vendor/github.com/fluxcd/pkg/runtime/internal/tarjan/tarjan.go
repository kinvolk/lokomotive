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

package tarjan

// Graph is a directed graph containing the vertex name and their Edges.
type Graph map[string]Edges

// Edges is a set of edges for a vertex.
type Edges map[string]struct{}

// SCC returns the strongly connected components of the given Graph.
func SCC(g Graph) [][]string {
	t := tarjan{
		g: g,

		indexTable: make(map[string]int, len(g)),
		lowLink:    make(map[string]int, len(g)),
		onStack:    make(map[string]bool, len(g)),
	}
	for v := range t.g {
		if t.indexTable[v] == 0 {
			t.strongConnect(v)
		}
	}
	return t.sccs
}

type tarjan struct {
	g Graph

	index      int
	indexTable map[string]int
	lowLink    map[string]int
	onStack    map[string]bool

	stack []string

	sccs [][]string
}

// strongConnect implements the pseudo-code from
// https://en.wikipedia.org/wiki/Tarjan%27s_strongly_connected_components_algorithm#The_algorithm_in_pseudocode
func (t *tarjan) strongConnect(v string) {
	// Set the depth index for v to the smallest unused index.
	t.index++
	t.indexTable[v] = t.index
	t.lowLink[v] = t.index
	t.stack = append(t.stack, v)
	t.onStack[v] = true

	// Consider successors of v.
	for w := range t.g[v] {
		if t.indexTable[w] == 0 {
			// Successor w has not yet been visited; recur on it.
			t.strongConnect(w)
			t.lowLink[v] = min(t.lowLink[v], t.lowLink[w])
		} else if t.onStack[w] {
			// Successor w is in stack s and hence in the current SCC.
			t.lowLink[v] = min(t.lowLink[v], t.indexTable[w])
		}
	}

	// If v is a root graph, pop the stack and generate an SCC.
	if t.lowLink[v] == t.indexTable[v] {
		// Start a new strongly connected component.
		var (
			scc []string
			w   string
		)
		for {
			w, t.stack = t.stack[len(t.stack)-1], t.stack[:len(t.stack)-1]
			t.onStack[w] = false
			// Add w to current strongly connected component.
			scc = append(scc, w)
			if w == v {
				break
			}
		}
		// Output the current strongly connected component.
		t.sccs = append(t.sccs, scc)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
