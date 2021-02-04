/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package controllers support control for the working of l8e operator.
package controllers

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/pingcap/errors"
	loger "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/kinvolk/lokomotive/cli/cmd/cluster"
	componentv1 "github.com/kinvolk/lokomotive/lokomotive-operator/api/v1"
)

const lokomotiveComponentFinalizer = "components.kinvolk.io/finalizer"

// LokomotiveComponentReconciler reconciles a LokomotiveComponent object.
type LokomotiveComponentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	event.DeleteEvent
}

//nolint:lll
// +kubebuilder:rbac:groups=components.kinvolk.io,resources=lokomotivecomponents,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.kinvolk.io,resources=lokomotivecomponents/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.kinvolk.io,resources=lokomotivecomponents/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LokomotiveComponent object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.1/pkg/reconcile
//nolint:funlen
func (r *LokomotiveComponentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("lokomotivecomponent", req.NamespacedName)

	// your logic here
	component := &componentv1.LokomotiveComponent{}

	// Since to run the lokomotive operator KUBECONFIG path should already be set,
	// so getting kubeconfig path like this is sufficient.
	kubeconfigFlag, found := os.LookupEnv("KUBECONFIG")
	if !found {
		return ctrl.Result{}, fmt.Errorf("env variable KUBECONFIG not set")
	}

	err := r.Get(ctx, req.NamespacedName, component)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("lokomotivecomponent resource not found. Ignoring since object must be deleted")

			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Lokomotivecomponent")

		return ctrl.Result{}, err
	}

	// Check if the LokomotiveComponent instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isLokomotiveComponentToBeDeleted := component.GetDeletionTimestamp() != nil
	if isLokomotiveComponentToBeDeleted { //nolint:nestif
		if contains(component.GetFinalizers(), lokomotiveComponentFinalizer) {
			// Run finalization logic for lokomotiveComponentFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.deleteComponent(component, kubeconfigFlag); err != nil {
				return ctrl.Result{}, err
			}

			// Remove lokomotiveComponentFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(component, lokomotiveComponentFinalizer)

			err := r.Update(ctx, component)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	// Add component and finalizer for this CR.
	if !contains(component.GetFinalizers(), lokomotiveComponentFinalizer) {
		if err := r.addComponent(log, component, kubeconfigFlag); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

//nolint:lll
func (r *LokomotiveComponentReconciler) deleteComponent(m *componentv1.LokomotiveComponent, kubeconfigFlag string) error {
	// Delete the component.
	componentName := m.Name
	configPath := m.Spec.Config[componentName+".lokocfg"]
	valuesPath := m.Spec.Config["lokocfg.vars"]

	contextLogger := loger.WithFields(loger.Fields{
		"operator": "applying component",
		"name":     componentName,
	})

	componentSlice := []string{componentName}

	options := cluster.ComponentDeleteOptions{
		Confirm:         true,
		DeleteNamespace: true,
		KubeconfigPath:  kubeconfigFlag,
		ConfigPath:      configPath,
		ValuesPath:      valuesPath,
	}

	if err := cluster.ComponentDelete(contextLogger, componentSlice, options); err != nil {
		return fmt.Errorf("deleting component: %v", err)
	}

	return nil
}

//nolint:lll
func (r *LokomotiveComponentReconciler) addComponent(log logr.Logger, m *componentv1.LokomotiveComponent, kubeconfigFlag string) error {
	// This will install component
	componentName := m.Name
	configPath := m.Spec.Config[componentName+".lokocfg"]
	valuesPath := m.Spec.Config["lokocfg.vars"]

	contextLogger := loger.WithFields(loger.Fields{
		"operator": "applying component",
		"name":     componentName,
	})

	componentSlice := []string{componentName}

	options := cluster.ComponentApplyOptions{
		KubeconfigPath: kubeconfigFlag,
		ConfigPath:     configPath,
		ValuesPath:     valuesPath,
	}

	if err := cluster.ComponentApply(contextLogger, componentSlice, options); err != nil {
		return fmt.Errorf("applying component: %v", err)
	}

	log.Info("Adding Finalizer for the LokomotiveComponent")
	controllerutil.AddFinalizer(m, lokomotiveComponentFinalizer)

	// Update CR
	err := r.Update(context.TODO(), m)
	if err != nil {
		log.Error(err, "Failed to update LokomotiveComponent with finalizer")

		return err
	}

	return nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}

	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *LokomotiveComponentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&componentv1.LokomotiveComponent{}).
		Complete(r)
}
