package k8sutil

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
)

// RolloutRestartDaemonSet is the programmatic equivalent of 'kubectl rollout restart'.
func RolloutRestartDaemonSet(dsi clientappsv1.DaemonSetInterface, name string) (int64, error) {
	ds, err := dsi.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("getting DaemonSet: %w", err)
	}

	if ds.Spec.Template.Annotations == nil {
		ds.Spec.Template.Annotations = map[string]string{}
	}

	ds.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().String()

	newDS, err := dsi.Update(context.TODO(), ds, metav1.UpdateOptions{})
	if err != nil {
		return 0, fmt.Errorf("updating DaemonSet: %w", err)
	}

	return newDS.Generation, nil
}

// RolloutRestartDeployment is the programmatic equivalent of 'kubectl rollout restart'.
func RolloutRestartDeployment(deployClient clientappsv1.DeploymentInterface, name string) (int64, error) {
	d, err := deployClient.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("getting Deployment: %w", err)
	}

	if d.Spec.Template.Annotations == nil {
		d.Spec.Template.Annotations = map[string]string{}
	}

	d.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().String()

	newDeploy, err := deployClient.Update(context.TODO(), d, metav1.UpdateOptions{})
	if err != nil {
		return 0, fmt.Errorf("updating Deployment: %w", err)
	}

	return newDeploy.Generation, nil
}

// daemonSetUpToDate checks if given DaemonSet converged.
func daemonSetUpToDate(ds *appsv1.DaemonSet, generation int64) (bool, error) {
	if ds.Spec.UpdateStrategy.Type != appsv1.RollingUpdateDaemonSetStrategyType {
		return true, fmt.Errorf("rollout status is only available for %s strategy type",
			appsv1.RollingUpdateStatefulSetStrategyType)
	}

	if generation > 0 && ds.Status.ObservedGeneration < generation {
		return false, nil
	}

	replicas := ds.Status.DesiredNumberScheduled

	if replicas == 0 {
		return false, nil
	}

	if ds.Status.NumberReady == replicas && ds.Status.UpdatedNumberScheduled == replicas {
		return true, nil
	}

	return false, nil
}

// WaitOptions holds optional arguments for WaitFor... functions group.
type WaitOptions struct {
	Generation int64
	Timeout    time.Duration
}

const (
	// WaitDefaultTimeout is default timeout after which watching for workload should return error.
	WaitDefaultTimeout = 5 * time.Minute
)

// WaitForDaemonSet waits until DaemonSet converges.
//
//nolint:dupl
func WaitForDaemonSet(dsi clientappsv1.DaemonSetInterface, name string, options WaitOptions) error {
	if dsi == nil {
		return fmt.Errorf("client must be set")
	}

	if name == "" {
		return fmt.Errorf("name must be set")
	}

	fieldSelector := fields.OneTermEqualSelector("metadata.name", name).String()
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fieldSelector

			return dsi.List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector

			return dsi.Watch(context.TODO(), options)
		},
	}

	watchF := func(e watch.Event) (bool, error) {
		switch t := e.Type; t { //nolint:exhaustive
		case watch.Added, watch.Modified:
			return daemonSetUpToDate(e.Object.(*appsv1.DaemonSet), options.Generation)
		case watch.Deleted:
			// We need to abort to avoid cases of recreation and not to silently watch the wrong (new) object
			return true, fmt.Errorf("object has been deleted")

		default:
			return true, fmt.Errorf("internal error: unexpected event %#v", e)
		}
	}

	// if the rollout isn't done yet, keep watching deployment status
	_, err := watchtools.UntilWithSync(context.TODO(), lw, &appsv1.DaemonSet{}, nil, watchF)
	if err != nil {
		return fmt.Errorf("waiting for DaemonSet to restart: %w", err)
	}

	return nil
}

// deploymentUpToDate checks if given deployed has converged.
func deploymentUpToDate(deploy *appsv1.Deployment, generation int64) (bool, error) {
	// Update has not been observed yet.
	if deploy.Status.ObservedGeneration < generation {
		return false, nil
	}

	var cond *appsv1.DeploymentCondition

	for i := range deploy.Status.Conditions {
		c := deploy.Status.Conditions[i]
		if c.Type == appsv1.DeploymentProgressing {
			cond = &c
		}
	}

	if cond != nil && cond.Reason == "ProgressDeadlineExceeded" {
		return false, fmt.Errorf("deployment %q exceeded its progress deadline", deploy.Name)
	}

	if deploy.Spec.Replicas != nil && deploy.Status.UpdatedReplicas < *deploy.Spec.Replicas {
		return false, nil
	}

	if deploy.Status.Replicas > deploy.Status.UpdatedReplicas {
		return false, nil
	}

	if deploy.Status.AvailableReplicas < deploy.Status.UpdatedReplicas {
		return false, nil
	}

	return true, nil
}

// WaitForDeployment waits for Deployment to converge.
//
//nolint:dupl
func WaitForDeployment(deployClient clientappsv1.DeploymentInterface, name string, options WaitOptions) error {
	if deployClient == nil {
		return fmt.Errorf("client must be set")
	}

	if name == "" {
		return fmt.Errorf("name must be set")
	}

	fieldSelector := fields.OneTermEqualSelector("metadata.name", name).String()
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fieldSelector

			return deployClient.List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector

			return deployClient.Watch(context.TODO(), options)
		},
	}

	watchF := func(e watch.Event) (bool, error) {
		switch t := e.Type; t { //nolint:exhaustive
		case watch.Added, watch.Modified:
			return deploymentUpToDate(e.Object.(*appsv1.Deployment), options.Generation)
		case watch.Deleted:
			// We need to abort to avoid cases of recreation and not to silently watch the wrong (new) object
			return true, fmt.Errorf("object has been deleted")

		default:
			return true, fmt.Errorf("internal error: unexpected event %#v", e)
		}
	}

	_, err := watchtools.UntilWithSync(context.TODO(), lw, &appsv1.Deployment{}, nil, watchF)
	if err != nil {
		return fmt.Errorf("waiting for Deployment to restart: %w", err)
	}

	return nil
}
