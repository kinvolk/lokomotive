package helmrelease

import (
	"context"
	"fmt"

	api "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/kinvolk/fluxlib/lib"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HelmReleaseConfig struct {
	c          client.Client
	kubeconfig []byte
}

type helmReleaseConfigOpt func(*HelmReleaseConfig)

var scheme *runtime.Scheme

func init() {
	scheme = runtime.NewScheme()
	_ = api.AddToScheme(scheme)
}

// helmReleaseCfg := lib.NewHelmReleaseConfig(
//     lib.WithKubeconfig(kc),
//     lib.WithFoobar(fb),
// )
func NewHelmReleaseConfig(fns ...helmReleaseConfigOpt) (*HelmReleaseConfig, error) {
	var ret HelmReleaseConfig

	for _, fn := range fns {
		fn(&ret)
	}

	var err error

	ret.c, err = lib.GetKubernetesClient(ret.kubeconfig, scheme)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func WithKubeconfig(kubeconfig []byte) helmReleaseConfigOpt {
	return func(hr *HelmReleaseConfig) {
		hr.kubeconfig = kubeconfig
	}
}

func (h *HelmReleaseConfig) Get(name, ns string) (*api.HelmRelease, error) {
	var got api.HelmRelease

	if err := h.c.Get(context.Background(), types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}, &got); err != nil {
		return nil, fmt.Errorf("getting HelmRelease: %w", err)
	}

	return &got, nil
}

func (h *HelmReleaseConfig) List(listOpts *client.ListOptions) (*api.HelmReleaseList, error) {
	var got api.HelmReleaseList

	if err := h.c.List(context.TODO(), &got, listOpts); err != nil {
		return nil, fmt.Errorf("listing HelmReleases: %w", err)
	}

	return &got, nil
}

func (h *HelmReleaseConfig) CreateOrUpdate(hr *api.HelmRelease) error {
	var got api.HelmRelease

	if err := h.c.Get(context.Background(), types.NamespacedName{
		Namespace: hr.GetNamespace(),
		Name:      hr.GetName(),
	}, &got); err != nil {
		if errors.IsNotFound(err) {
			// Create the object since it does not exists.
			if err := h.c.Create(context.Background(), hr); err != nil {
				return fmt.Errorf("creating HelmRelease: %w", err)
			}

			return nil
		}

		return fmt.Errorf("looking up HelmRelease: %w", err)
	}

	hr.ResourceVersion = got.ResourceVersion

	if err := h.c.Update(context.Background(), hr); err != nil {
		return fmt.Errorf("updating HelmRelease: %w", err)
	}

	return nil
}

func (h *HelmReleaseConfig) Delete(hr *api.HelmRelease) error {
	if err := h.c.Delete(context.Background(), hr); err != nil {
		return fmt.Errorf("deleting HelmRelease: %w", err)
	}

	return nil
}
