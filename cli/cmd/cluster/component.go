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

package cluster

import (
	// Register a component by adding an anonymous import
	_ "github.com/kinvolk/lokomotive/pkg/components/aws-ebs-csi-driver"
	_ "github.com/kinvolk/lokomotive/pkg/components/cert-manager"
	_ "github.com/kinvolk/lokomotive/pkg/components/cluster-autoscaler"
	_ "github.com/kinvolk/lokomotive/pkg/components/contour"
	_ "github.com/kinvolk/lokomotive/pkg/components/dex"
	_ "github.com/kinvolk/lokomotive/pkg/components/external-dns"
	_ "github.com/kinvolk/lokomotive/pkg/components/flatcar-linux-update-operator"
	_ "github.com/kinvolk/lokomotive/pkg/components/gangway"
	_ "github.com/kinvolk/lokomotive/pkg/components/httpbin"
	_ "github.com/kinvolk/lokomotive/pkg/components/inspektor-gadget"
	_ "github.com/kinvolk/lokomotive/pkg/components/istio-operator"
	_ "github.com/kinvolk/lokomotive/pkg/components/linkerd"
	_ "github.com/kinvolk/lokomotive/pkg/components/metallb"
	_ "github.com/kinvolk/lokomotive/pkg/components/metrics-server"
	_ "github.com/kinvolk/lokomotive/pkg/components/openebs-operator"
	_ "github.com/kinvolk/lokomotive/pkg/components/openebs-storage-class"
	_ "github.com/kinvolk/lokomotive/pkg/components/prometheus-operator"
	_ "github.com/kinvolk/lokomotive/pkg/components/rook"
	_ "github.com/kinvolk/lokomotive/pkg/components/rook-ceph"
	_ "github.com/kinvolk/lokomotive/pkg/components/velero"
	_ "github.com/kinvolk/lokomotive/pkg/components/web-ui"

	"github.com/kinvolk/lokomotive/pkg/components"
)

// AvailableComponents returns list of valid component names.
func AvailableComponents() []string {
	return components.ListNames()
}
