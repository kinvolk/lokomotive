package components

import (
	"fmt"
)

// components is the map of registered components
var components map[string]Component

func init() {
	components = make(map[string]Component)
}

func Register(name string, obj Component) {
	if _, exists := components[name]; exists {
		panic(fmt.Sprintf("component with name %q registered already", name))
	}
	components[name] = obj
}

func ListNames() []string {
	var componentList []string
	for name := range components {
		componentList = append(componentList, name)
	}
	return componentList
}

func List() []Component {
	var componentList []Component
	for _, component := range components {
		componentList = append(componentList, component)
	}
	return componentList
}

func Get(name string) (Component, error) {
	component, exists := components[name]
	if !exists {
		return nil, fmt.Errorf("no component with name %q found", name)
	}
	return component, nil
}
