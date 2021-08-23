package sourcecontroller

import (
	"context"
	"fmt"

	api "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/kinvolk/fluxlib/lib"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GitRepoConfig struct {
	c          client.Client
	kubeconfig []byte
}

type gitRespoConfigOpt func(*GitRepoConfig)

var scheme *runtime.Scheme

func init() {
	scheme = runtime.NewScheme()
	_ = api.AddToScheme(scheme)
}

// gitRepoCfg := lib.NewGitRepoConfig(
//     lib.WithGitRepository(gr),
//     lib.WithKubeconfig(kc),
//     lib.WithFoobar(fb),
// )
func NewGitRepoConfig(fns ...gitRespoConfigOpt) (*GitRepoConfig, error) {
	var ret GitRepoConfig

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

func WithKubeconfig(kubeconfig []byte) gitRespoConfigOpt {
	return func(gr *GitRepoConfig) {
		gr.kubeconfig = kubeconfig
	}
}

func (g *GitRepoConfig) Get(name, ns string) (*api.GitRepository, error) {
	var got api.GitRepository

	if err := g.c.Get(context.Background(), types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}, &got); err != nil {
		return nil, fmt.Errorf("getting GitRepository: %w", err)
	}

	return &got, nil
}

func (g *GitRepoConfig) List(listOpts *client.ListOptions) (*api.GitRepositoryList, error) {
	var got api.GitRepositoryList

	if err := g.c.List(context.TODO(), &got, listOpts); err != nil {
		return nil, fmt.Errorf("listing GitRepositories: %w", err)
	}

	return &got, nil
}

func (g *GitRepoConfig) CreateOrUpdate(gr *api.GitRepository) error {
	var got api.GitRepository

	if err := g.c.Get(context.Background(), types.NamespacedName{
		Namespace: gr.GetNamespace(),
		Name:      gr.GetName(),
	}, &got); err != nil {
		if errors.IsNotFound(err) {
			// Create the object since it does not exists.
			if err := g.c.Create(context.Background(), gr); err != nil {
				return fmt.Errorf("creating GitRepository: %w", err)
			}

			return nil
		}

		return fmt.Errorf("looking up GitRepository: %w", err)
	}

	gr.ResourceVersion = got.ResourceVersion

	if err := g.c.Update(context.Background(), gr); err != nil {
		return fmt.Errorf("updating GitRepository: %w", err)
	}

	return nil
}

func (g *GitRepoConfig) Delete(gr *api.GitRepository) error {
	if err := g.c.Delete(context.Background(), gr); err != nil {
		return fmt.Errorf("deleting HelmRelease: %w", err)
	}

	return nil
}
