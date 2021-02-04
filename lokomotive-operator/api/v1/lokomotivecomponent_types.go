/*
Copyright 2021.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LokomotiveComponentSpec defines the desired state of LokomotiveComponent
type LokomotiveComponentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	// Config is the lokomotive config passed for components.
	Config map[string]string `json:"config"`
}

// LokomotiveComponentStatus defines the observed state of LokomotiveComponent
type LokomotiveComponentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.metadata.namespace`

// LokomotiveComponent is the Schema for the lokomotivecomponents API
type LokomotiveComponent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LokomotiveComponentSpec   `json:"spec,omitempty"`
	Status LokomotiveComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LokomotiveComponentList contains a list of LokomotiveComponent
type LokomotiveComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LokomotiveComponent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LokomotiveComponent{}, &LokomotiveComponentList{})
}
