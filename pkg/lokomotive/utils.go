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

package lokomotive

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

// askForConfirmation asks the user to confirm an action.
// It prints the message and then asks the user to type "yes" or "no".
// If the user types "yes" the function returns true, otherwise it returns
// false.
func askForConfirmation(message string) (bool, error) {
	var input string

	fmt.Printf("%s [type \"yes\" to continue]: ", message)

	_, err := fmt.Scanln(&input)
	if err != nil {
		return false, err
	}

	return (input == "yes"), nil
}

// expandKubeconfigPath tries to expand ~ in the given kubeconfig path.
// However, if that fails, it just returns original path as the best effort.
func expandKubeconfigPath(path string) string {
	if expandedPath, err := homedir.Expand(path); err == nil {
		return expandedPath
	}

	// homedir.Expand is too restrictive for the ~ prefix,
	// i.e., it errors on "~somepath" which is a valid path,
	// so just return the original path.
	return path
}

// getKubeconfig finds the kubeconfig to be used. Precedence takes a specified
// flag or environment variable. Then the asset directory of the cluster is searched
// and finally the global default value is used. This cannot be done in Viper
// because we need the other values from Viper to find the asset directory.
func getKubeconfig(assetDir string) string {
	kubeconfig := viper.GetString("kubeconfig")
	if kubeconfig != "" {
		return expandKubeconfigPath(kubeconfig)
	}

	if assetDir != "" {
		return expandKubeconfigPath(assetsKubeconfig(assetDir))
	}

	return expandKubeconfigPath("~/.kube/config")
}

func assetsKubeconfig(assetDir string) string {
	return filepath.Join(assetDir, "cluster-assets", "auth", "kubeconfig")
}

// doesKubeconfigExist checks if the kubeconfig provided by user exists
func doesKubeconfigExist(assetDir string) error {
	var err error

	kubeconfig := getKubeconfig(assetDir)
	if _, err = os.Stat(kubeconfig); os.IsNotExist(err) {
		return fmt.Errorf("kubeconfig %q not found", kubeconfig)
	}

	return err
}

func deleteNS(ns string, kubeconfig string) error {
	cs, err := k8sutil.NewClientset(kubeconfig)
	if err != nil {
		return err
	}
	// Delete the manually created namespace which was not created by helm.
	if err = cs.CoreV1().Namespaces().Delete(ns, &metav1.DeleteOptions{}); err != nil {
		// Ignore error when the namespace does not exist.
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	return nil
}
