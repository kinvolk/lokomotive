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
package k8sutil

import (
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// daemonSetReady checks if Pods of a given DaemonSet all ready.
// If DaemonSet does not exist, false will be returned.
// If number of replicas for DaemonSet is 0, false will be returned as well.
func DaemonSetReady(client kubernetes.Interface, ns, name string) (bool, error) {
	ds, err := client.AppsV1().DaemonSets(ns).Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	replicas := ds.Status.DesiredNumberScheduled

	return replicas != 0 && ds.Status.NumberReady == replicas, nil
}
