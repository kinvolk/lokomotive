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

package platform

import (
	"github.com/kinvolk/lokomotive/pkg/terraform"
	"github.com/kinvolk/lokomotive/pkg/version"
)

const (
	// AKS represents an AKS cluster.
	AKS = "aks"
	// AWS represents an AWS cluster.
	AWS = "aws"
	// BareMetal represents a bare metal cluster.
	BareMetal = "baremetal"
	// Packet represents a Packet cluster.
	Packet = "packet"
)

// CommonControlPlaneCharts defines a list of control plane Helm charts to be deployed for all
// platforms.
var CommonControlPlaneCharts = []string{
	"calico",
	"kube-apiserver",
	"kubernetes",
	"pod-checkpointer",
}

// Cluster describes a Lokomotive cluster.
type Cluster interface {
	// AssetDir returns the path to the Lokomotive assets directory.
	AssetDir() string
	// ControlPlaneCharts returns a list of Helm charts which compose the k8s control plane.
	ControlPlaneCharts() []string
	// Managed returns true if the cluster uses a managed platform (e.g. AKS).
	Managed() bool
	// Nodes returns the total number of nodes for the cluster. This is the total number of nodes
	// including all controller nodes and all worker nodes from all worker pools.
	Nodes() int
	// TerraformExecutionPlan returns a list of terraform.ExecutionStep representing steps which
	// should be executed to get a working cluster on a platform. The execution plan is used during
	// cluster creation only - when destroying a cluster, a simple `terraform destroy` is always
	// executed.
	//
	// The commands specified in the Args field of each TerraformExecutionStep are passed as
	// arguments to the `terraform` binary and are executed in order.
	// `apply` operations should be followed by `-auto-approve` to skip interactive prompts.
	TerraformExecutionPlan() []terraform.ExecutionStep
	// TerraformRootModule returns a string representing the contens of the root Terraform module
	// which should be used for cluster operations.
	TerraformRootModule() string
}

// AppendVersionTag appends the lokoctl-version tag to a given tags map.
func AppendVersionTag(tags *map[string]string) {
	if tags == nil {
		return
	}

	if *tags == nil {
		*tags = make(map[string]string)
	}

	if version.Version != "" {
		(*tags)["lokoctl-version"] = version.Version
	}
}
