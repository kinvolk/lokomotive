package components

// Metadata is a struct which represents basic information about the component.
// It may contain information like name, version, dependencies, namespace, source etc.
type Metadata struct {
	Namespace string
	Helm      *HelmMetadata
}

// HelmMetadata stores Helm-related information about a component that is needed when managing component using Helm.
type HelmMetadata struct {
	Wait bool
}
