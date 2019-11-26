package metallb

import (
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
	"github.com/pkg/errors"
)

const name = "metallb"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	ControllerNodeSelectors map[string]string `hcl:"controller_node_selectors,optional"`
	SpeakerNodeSelectors    map[string]string `hcl:"speaker_node_selectors,optional"`
	ControllerTolerations   []util.Toleration `hcl:"controller_toleration,block"`
	SpeakerTolerations      []util.Toleration `hcl:"speaker_toleration,block"`
	ServiceMonitor          bool              `hcl:"service_monitor,optional"`

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

	controllerStr, err := util.RenderTemplate(deploymentController, c)
	if err != nil {
		return nil, errors.Wrap(err, "render template failed")
	}

	speakerStr, err := util.RenderTemplate(daemonsetSpeaker, c)
	if err != nil {
		return nil, errors.Wrap(err, "render template failed")
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
	}

	// Create service and service monitor for Prometheus to scrape metrics
	if c.ServiceMonitor {
		rendered["service.yaml"] = service
		rendered["service-monitor.yaml"] = serviceMonitor
	}

	return rendered, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Namespace: "metallb-system",
	}
}
