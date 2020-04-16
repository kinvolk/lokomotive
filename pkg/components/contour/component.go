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

package contour

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/kinvolk/lokomotive/pkg/util"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/util/walkers"
)

const name = "contour"

func init() {
	components.Register(name, newComponent())
}

// IngressHosts field is added in order to make contour work with ExternalDNS component.
// Values provided for IngressHosts is used as value for the annotation `external-dns.alpha.kubernetes.io/hostname`
// This annotation is added to Envoy service.
type component struct {
	ServiceMonitor bool `hcl:"service_monitor,optional"`
	// IngressHosts field is added in order to make contour work with ExternalDNS component.
	// Values provided for IngressHosts is used as value for the annotation `external-dns.alpha.kubernetes.io/hostname`.
	// This annotation is added to Envoy Service, in order for ExternalDNS to create DNS entries.
	// This solution is a workaround for projectcontour/contour#403
	// More details regarding this workaround and other solutions is captured in
	// https://github.com/kinvolk/PROJECT-Lokomotive-Kubernetes/issues/474
	IngressHosts []string `hcl:"ingress_hosts,optional"`

	// IngressHostsRaw is not accessible to the user
	IngressHostsRaw string
}

func newComponent() *component {
	return &component{}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{
			components.HCLDiagConfigBodyNil,
		}
	}
	if err := gohcl.DecodeBody(*configBody, evalContext, c); err != nil {
		return err
	}

	return nil
}

func (c *component) RenderManifests() (map[string]string, error) {
	ret := make(map[string]string)

	walk := walkers.DumpingWalker(ret, ".yaml")
	if err := assets.Assets.WalkFiles(fmt.Sprintf("/components/%s/%s", name, name), walk); err != nil {
		return nil, errors.Wrap(err, "failed to walk assets")
	}

	// Create service and service monitor for Prometheus to scrape metrics
	if c.ServiceMonitor {
		if err := assets.Assets.WalkFiles(fmt.Sprintf("/components/%s/manifests-metrics", name), walk); err != nil {
			return nil, errors.Wrap(err, "failed to walk assets")
		}
	}
	// To store the comma separated string representation of IngressHosts
	c.IngressHostsRaw = strings.Join(c.IngressHosts, ",")
	// Parse envoy service template.
	envoyServiceStr, err := util.RenderTemplate(envoyServiceTmpl, c)
	if err != nil {
		return nil, errors.Wrap(err, "render template failed")
	}
	ret["02-service-envoy.yaml"] = envoyServiceStr

	return ret, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name:      name,
		Namespace: "projectcontour",
	}
}
