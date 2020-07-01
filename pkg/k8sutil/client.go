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
	"fmt"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClientset creates new Kubernetes Client set object from the contents
// of the given kubeconfig file.
func NewClientset(data []byte) (*kubernetes.Clientset, error) {
	c, err := clientcmd.NewClientConfigFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("creating client config failed: %w", err)
	}

	restConfig, err := c.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("converting client config to rest client config failed: %w", err)
	}

	return kubernetes.NewForConfig(restConfig)
}
