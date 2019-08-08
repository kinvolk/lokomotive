package metallb

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
)

const name = "metallb"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	ControllerNodeSelectors map[string]string `hcl:"controller_node_selectors,optional"`
	SpeakerNodeSelectors    map[string]string `hcl:"speaker_node_selectors,optional"`
	ControllerTolerations   []toleration      `hcl:"controller_toleration,block"`
	SpeakerTolerations      []toleration      `hcl:"speaker_toleration,block"`
}

type toleration struct {
	Key               string `hcl:"key,optional" json:"key,omitempty"`
	Effect            string `hcl:"effect,optional" json:"effect,omitempty"`
	Operator          string `hcl:"operator,optional" json:"operator,omitempty"`
	Value             string `hcl:"value,optional" json:"value,omitempty"`
	TolerationSeconds string `hcl:"toleration_seconds,optional" json:"toleration_seconds,omitempty"`
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

// renderTolerations takes a list of tolerations.
// It returns a json string and an error if any.
func renderTolerations(t []toleration) (string, error) {
	if len(t) == 0 {
		return "", nil
	}

	b, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (c *component) RenderManifests() (map[string]string, error) {
	st, err := renderTolerations(c.SpeakerTolerations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal speaker tolerations")
	}

	ct, err := renderTolerations(c.ControllerTolerations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal controller tolerations")
	}

	cv := struct {
		ControllerNodeSelectors map[string]string
		SpeakerNodeSelectors    map[string]string
		SpeakerTolerations      string
		ControllerTolerations   string
	}{
		ControllerNodeSelectors: c.ControllerNodeSelectors,
		SpeakerNodeSelectors:    c.SpeakerNodeSelectors,
		SpeakerTolerations:      st,
		ControllerTolerations:   ct,
	}

	tmpl, err := template.New("controller").Parse(deploymentController)
	if err != nil {
		return nil, errors.Wrap(err, "parse template failed")
	}

	var controllerBuf bytes.Buffer
	if err := tmpl.Execute(&controllerBuf, cv); err != nil {
		return nil, errors.Wrap(err, "execute template failed")
	}

	tmpl, err = template.New("speaker").Parse(daemonsetSpeaker)
	if err != nil {
		return nil, errors.Wrap(err, "parse template failed")
	}

	var speakerBuf bytes.Buffer
	if err := tmpl.Execute(&speakerBuf, cv); err != nil {
		return nil, errors.Wrap(err, "execute template failed")
	}

	return map[string]string{
		"namespace.yaml":                                    namespace,
		"service-account-controller.yaml":                   serviceAccountController,
		"service-account-speaker.yaml":                      serviceAccountSpeaker,
		"clusterrole-metallb-system-controller.yaml":        clusterRoleMetallbSystemController,
		"clusterrole-metallb-System-speaker.yaml":           clusterRoleMetallbSystemSpeaker,
		"role-config-watcher.yaml":                          roleConfigWatcher,
		"clusterrolebinding-metallb-system-controller.yaml": clusterRoleBindingMetallbSystemController,
		"clusterrolebinding-metallb-system-speaker.yaml":    clusterRoleBindingMetallbSystemSpeaker,
		"rolebinding-config-watcher.yaml":                   roleBindingConfigWatcher,
		"deployment-controller.yaml":                        controllerBuf.String(),
		"daemonset-speaker.yaml":                            speakerBuf.String(),
		"psp-metallb-speaker.yaml":                          pspMetallbSpeaker,
	}, nil
}

func (c *component) Install(kubeconfig string) error {
	return util.Install(c, kubeconfig)
}
