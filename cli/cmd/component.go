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

package cmd

import (
	"github.com/spf13/cobra"

	// Register a component by adding an anonymous import
	_ "github.com/kinvolk/lokoctl/pkg/components/calico-hostendpoint-controller"
	_ "github.com/kinvolk/lokoctl/pkg/components/cert-manager"
	_ "github.com/kinvolk/lokoctl/pkg/components/cluster-autoscaler"
	_ "github.com/kinvolk/lokoctl/pkg/components/contour"
	_ "github.com/kinvolk/lokoctl/pkg/components/dex"
	_ "github.com/kinvolk/lokoctl/pkg/components/external-dns"
	_ "github.com/kinvolk/lokoctl/pkg/components/flatcar-linux-update-operator"
	_ "github.com/kinvolk/lokoctl/pkg/components/gangway"
	_ "github.com/kinvolk/lokoctl/pkg/components/httpbin"
	_ "github.com/kinvolk/lokoctl/pkg/components/metallb"
	_ "github.com/kinvolk/lokoctl/pkg/components/metrics-server"
	_ "github.com/kinvolk/lokoctl/pkg/components/openebs-operator"
	_ "github.com/kinvolk/lokoctl/pkg/components/openebs-storage-class"
	_ "github.com/kinvolk/lokoctl/pkg/components/prometheus-operator"
	_ "github.com/kinvolk/lokoctl/pkg/components/rook"
	_ "github.com/kinvolk/lokoctl/pkg/components/rook-ceph"
	_ "github.com/kinvolk/lokoctl/pkg/components/velero"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Install Lokomotive components",
}

func init() {
	rootCmd.AddCommand(componentCmd)
}
