// Copyright 2021 The Lokomotive Authors
// Copyright 2016 The Kubernetes Authors
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

// RolloutTimeout is the timeout to wait for a daemonset or deployment to rollout changes.
const RolloutTimeout = 15 * time.Minute

// RolloutDaemonSet runs rolloutRestartDaemonSet and will wait for the DaemonSet to be fully rolled out.
func RolloutDaemonSet(ctx context.Context, dsi clientappsv1.DaemonSetInterface, name string) error {
	generation, err := rolloutRestartDaemonSet(ctx, dsi, name)
	if err != nil {
		return fmt.Errorf("restarting: %w", err)
	}

	options := WaitOptions{
		Generation: generation,
		Timeout:    RolloutTimeout,
	}

	if err := waitForDaemonSet(ctx, dsi, name, options); err != nil {
		return fmt.Errorf("waiting for DaemonSet to converge: %w", err)
	}

	return nil
}

// RolloutDeployment runs rolloutRestartDeployment and will wait for the DaemonSet to be fully rolled out.
func RolloutDeployment(ctx context.Context, deployClient clientappsv1.DeploymentInterface, name string) error {
	generation, err := rolloutRestartDeployment(ctx, deployClient, name)
	if err != nil {
		return fmt.Errorf("restarting: %w", err)
	}

	options := WaitOptions{
		Generation: generation,
		Timeout:    RolloutTimeout,
	}

	if err := waitForDeployment(ctx, deployClient, name, options); err != nil {
		return fmt.Errorf("waiting for Deployment to converge: %w", err)
	}

	return nil
}

// rolloutRestartDaemonSet is the programmatic equivalent of 'kubectl rollout restart'.
func rolloutRestartDaemonSet(ctx context.Context, dsi clientappsv1.DaemonSetInterface, name string) (int64, error) {
	ds, err := dsi.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("getting DaemonSet: %w", err)
	}

	if ds.Spec.Template.Annotations == nil {
		ds.Spec.Template.Annotations = map[string]string{}
	}

	// We mimic what "kubectl rollout restart" does and set the restartedAt
	// annotation. This change in the object causes a restart.
	ds.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	newDS, err := dsi.Update(ctx, ds, metav1.UpdateOptions{})
	if err != nil {
		return 0, fmt.Errorf("updating DaemonSet: %w", err)
	}

	return newDS.Generation, nil
}

// rolloutRestartDeployment is the programmatic equivalent of 'kubectl rollout restart'.
func rolloutRestartDeployment(ctx context.Context,
	deployClient clientappsv1.DeploymentInterface, name string) (int64, error) {
	d, err := deployClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("getting Deployment: %w", err)
	}

	if d.Spec.Template.Annotations == nil {
		d.Spec.Template.Annotations = map[string]string{}
	}

	// We mimic what "kubectl rollout restart" does and set the restartedAt
	// annotation. This change in the spec causes all Daemonset Pods to be
	// recreated.
	d.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	newDeploy, err := deployClient.Update(ctx, d, metav1.UpdateOptions{})
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

// waitForDaemonSet waits until DaemonSet converges.
//
//nolint:dupl
func waitForDaemonSet(ctx context.Context,
	dsi clientappsv1.DaemonSetInterface, name string, options WaitOptions) error {
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

			return dsi.List(ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector

			return dsi.Watch(ctx, options)
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

	// If the rollout isn't done yet, keep watching deployment status.
	if _, err := watchtools.UntilWithSync(ctx, lw, &appsv1.DaemonSet{}, nil, watchF); err != nil {
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

// waitForDeployment waits for Deployment to converge.
//
//nolint:dupl
func waitForDeployment(ctx context.Context,
	deployClient clientappsv1.DeploymentInterface, name string, options WaitOptions) error {
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

			return deployClient.List(ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector

			return deployClient.Watch(ctx, options)
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

	if _, err := watchtools.UntilWithSync(ctx, lw, &appsv1.Deployment{}, nil, watchF); err != nil {
		return fmt.Errorf("waiting for Deployment to restart: %w", err)
	}

	return nil
}
