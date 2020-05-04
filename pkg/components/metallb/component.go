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

package metallb

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

const name = "metallb"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	AddressPools            map[string][]string `hcl:"address_pools"`
	ControllerNodeSelectors map[string]string   `hcl:"controller_node_selectors,optional"`
	SpeakerNodeSelectors    map[string]string   `hcl:"speaker_node_selectors,optional"`
	ControllerTolerations   []util.Toleration   `hcl:"controller_toleration,block"`
	SpeakerTolerations      []util.Toleration   `hcl:"speaker_toleration,block"`
	ServiceMonitor          bool                `hcl:"service_monitor,optional"`

	ControllerTolerationsJSON string
	SpeakerTolerationsJSON    string
}

func newComponent() *component {
	return &component{}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	// Here are `nodeSelectors` and `tolerations` that are set by upstream. To make sure that we
	// don't miss them out we set them manually here. We cannot make these changes in the template
	// because we have parameterized these fields.
	if c.SpeakerNodeSelectors == nil {
		c.SpeakerNodeSelectors = map[string]string{}
	}
	// MetalLB only supports Linux, so force this selector, even if it's already specified by the
	// user.
	c.SpeakerNodeSelectors["beta.kubernetes.io/os"] = "linux"

	if c.ControllerNodeSelectors == nil {
		c.ControllerNodeSelectors = map[string]string{}
	}
	c.ControllerNodeSelectors["beta.kubernetes.io/os"] = "linux"
	c.ControllerNodeSelectors["node.kubernetes.io/master"] = ""

	c.ControllerTolerations = append(c.SpeakerTolerations, util.Toleration{
		Effect: "NoSchedule",
		Key:    "node-role.kubernetes.io/master",
	})

	t, err := util.RenderTolerations(c.SpeakerTolerations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal speaker tolerations")
	}
	c.SpeakerTolerationsJSON = t

	t, err = util.RenderTolerations(c.ControllerTolerations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal controller tolerations")
	}
	c.ControllerTolerationsJSON = t

	controllerStr, err := template.Render(deploymentController, c)
	if err != nil {
		return nil, errors.Wrap(err, "render template failed")
	}

	speakerStr, err := template.Render(daemonsetSpeaker, c)
	if err != nil {
		return nil, errors.Wrap(err, "render template failed")
	}

	configMapStr, err := template.Render(configMap, c)
	if err != nil {
		return nil, errors.Wrap(err, "rendering ConfigMap template failed")
	}

	rendered := map[string]string{
		"namespace.yaml":                                    namespace,
		"service-account-controller.yaml":                   serviceAccountController,
		"service-account-speaker.yaml":                      serviceAccountSpeaker,
		"clusterrole-metallb-system-controller.yaml":        clusterRoleMetallbSystemController,
		"clusterrole-metallb-System-speaker.yaml":           clusterRoleMetallbSystemSpeaker,
		"role-config-watcher.yaml":                          roleConfigWatcher,
		"clusterrolebinding-metallb-system-controller.yaml": clusterRoleBindingMetallbSystemController,
		"clusterrolebinding-metallb-system-speaker.yaml":    clusterRoleBindingMetallbSystemSpeaker,
		"rolebinding-config-watcher.yaml":                   roleBindingConfigWatcher,
		"deployment-controller.yaml":                        controllerStr,
		"daemonset-speaker.yaml":                            speakerStr,
		"psp-metallb-speaker.yaml":                          pspMetallbSpeaker,
		"configmap.yaml":                                    configMapStr,
	}

	// Create service and service monitor for Prometheus to scrape metrics
	if c.ServiceMonitor {
		rendered["service.yaml"] = service
		rendered["service-monitor.yaml"] = serviceMonitor
		rendered["grafana-dashboard.yaml"] = grafanaDashboard
		rendered["grafana-alertmanager-rule.yaml"] = metallbPrometheusRule
	}

	return rendered, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name:      name,
		Namespace: "metallb-system",
	}
}
