package components

import (
	"fmt"

	"github.com/kinvolk/lokoctl/pkg/component"
	"github.com/kinvolk/lokoctl/pkg/component/ingressnginx"
	"github.com/kinvolk/lokoctl/pkg/component/networkpolicy"
)

var components = make(map[string]component.Interface)

func Register(c component.Interface) error {
	key := c.Name()
	if _, ok := components[key]; ok {
		return fmt.Errorf("component %q is already registered", key)
	}

	components[c.Name()] = c

	return nil
}

func List() []string {
	keys := make([]string, 0, len(components))
	for k := range components {
		keys = append(keys, k)
	}

	return keys
}

func Get(name string) (component.Interface, error) {
	if c, ok := components[name]; ok {
		return c, nil
	}

	return nil, fmt.Errorf("no such component %q", name)
}

func init() {
	// register your components here

	ig := ingressnginx.New()
	Register(ig)

	dnp := networkpolicy.New()
	Register(dnp)
}
