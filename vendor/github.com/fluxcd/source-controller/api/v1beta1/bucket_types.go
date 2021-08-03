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
	"github.com/fluxcd/pkg/apis/meta"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// BucketKind is the string representation of a Bucket.
	BucketKind = "Bucket"
)

// BucketSpec defines the desired state of an S3 compatible bucket
type BucketSpec struct {
	// The S3 compatible storage provider name, default ('generic').
	// +kubebuilder:validation:Enum=generic;aws
	// +kubebuilder:default:=generic
	// +optional
	Provider string `json:"provider,omitempty"`

	// The bucket name.
	// +required
	BucketName string `json:"bucketName"`

	// The bucket endpoint address.
	// +required
	Endpoint string `json:"endpoint"`

	// Insecure allows connecting to a non-TLS S3 HTTP endpoint.
	// +optional
	Insecure bool `json:"insecure,omitempty"`

	// The bucket region.
	// +optional
	Region string `json:"region,omitempty"`

	// The name of the secret containing authentication credentials
	// for the Bucket.
	// +optional
	SecretRef *meta.LocalObjectReference `json:"secretRef,omitempty"`

	// The interval at which to check for bucket updates.
	// +required
	Interval metav1.Duration `json:"interval"`

	// The timeout for download operations, defaults to 20s.
	// +kubebuilder:default="20s"
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// Ignore overrides the set of excluded patterns in the .sourceignore format
	// (which is the same as .gitignore). If not provided, a default will be used,
	// consult the documentation for your version to find out what those are.
	// +optional
	Ignore *string `json:"ignore,omitempty"`

	// This flag tells the controller to suspend the reconciliation of this source.
	// +optional
	Suspend bool `json:"suspend,omitempty"`
}

const (
	GenericBucketProvider string = "generic"
	AmazonBucketProvider  string = "aws"
)

// BucketStatus defines the observed state of a bucket
type BucketStatus struct {
	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds the conditions for the Bucket.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// URL is the download link for the artifact output of the last Bucket sync.
	// +optional
	URL string `json:"url,omitempty"`

	// Artifact represents the output of the last successful Bucket sync.
	// +optional
	Artifact *Artifact `json:"artifact,omitempty"`

	meta.ReconcileRequestStatus `json:",inline"`
}

const (
	// BucketOperationSucceedReason represents the fact that the bucket listing and
	// download operations succeeded.
	BucketOperationSucceedReason string = "BucketOperationSucceed"

	// BucketOperationFailedReason represents the fact that the bucket listing or
	// download operations failed.
	BucketOperationFailedReason string = "BucketOperationFailed"
)

// BucketProgressing resets the conditions of the Bucket to metav1.Condition of
// type meta.ReadyCondition with status 'Unknown' and meta.ProgressingReason
// reason and message. It returns the modified Bucket.
func BucketProgressing(bucket Bucket) Bucket {
	bucket.Status.ObservedGeneration = bucket.Generation
	bucket.Status.URL = ""
	bucket.Status.Conditions = []metav1.Condition{}
	meta.SetResourceCondition(&bucket, meta.ReadyCondition, metav1.ConditionUnknown, meta.ProgressingReason, "reconciliation in progress")
	return bucket
}

// BucketReady sets the given Artifact and URL on the Bucket and sets the
// meta.ReadyCondition to 'True', with the given reason and message. It returns
// the modified Bucket.
func BucketReady(bucket Bucket, artifact Artifact, url, reason, message string) Bucket {
	bucket.Status.Artifact = &artifact
	bucket.Status.URL = url
	meta.SetResourceCondition(&bucket, meta.ReadyCondition, metav1.ConditionTrue, reason, message)
	return bucket
}

// BucketNotReady sets the meta.ReadyCondition on the Bucket to 'False', with
// the given reason and message. It returns the modified Bucket.
func BucketNotReady(bucket Bucket, reason, message string) Bucket {
	meta.SetResourceCondition(&bucket, meta.ReadyCondition, metav1.ConditionFalse, reason, message)
	return bucket
}

// BucketReadyMessage returns the message of the metav1.Condition of type
// meta.ReadyCondition with status 'True' if present, or an empty string.
func BucketReadyMessage(bucket Bucket) string {
	if c := apimeta.FindStatusCondition(bucket.Status.Conditions, meta.ReadyCondition); c != nil {
		if c.Status == metav1.ConditionTrue {
			return c.Message
		}
	}
	return ""
}

// GetArtifact returns the latest artifact from the source if present in the
// status sub-resource.
func (in *Bucket) GetArtifact() *Artifact {
	return in.Status.Artifact
}

// GetStatusConditions returns a pointer to the Status.Conditions slice
func (in *Bucket) GetStatusConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

// GetInterval returns the interval at which the source is updated.
func (in *Bucket) GetInterval() metav1.Duration {
	return in.Spec.Interval
}

// +genclient
// +genclient:Namespaced
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.url`
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""

// Bucket is the Schema for the buckets API
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketSpec   `json:"spec,omitempty"`
	Status BucketStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BucketList contains a list of Bucket
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bucket{}, &BucketList{})
}
