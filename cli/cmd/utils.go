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

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	kubeconfigEnvVariable = "KUBECONFIG"
	defaultKubeconfigPath = "~/.kube/config"
)

func getKubeconfig(assetDir string) ([]byte, error) {
	path := kubeconfigPath(assetDir)

	if expandedPath, err := homedir.Expand(path); err == nil {
		path = expandedPath
	}

	// homedir.Expand is too restrictive for the ~ prefix,
	// i.e., it errors on "~somepath" which is a valid path,
	// so just read from the original path.
	return ioutil.ReadFile(path) // #nosec G304
}

// getKubeconfig finds the kubeconfig to be used. The precedence is the following:
// - --kubeconfig-file flag OR KUBECONFIG_FILE environment variable (the latter
// is a side-effect of cobra/viper and should NOT be documented because it's
// confusing).
// - Asset directory from cluster configuration.
// - KUBECONFIG environment variable.
// - ~/.kube/config path, which is the default for kubectl.

// kubeconfigPath returns a path to a kubeconfig file using the following order of precedence:
// - The value provided via the --kubeconfig-file flag or the KUBECONFIG_FILE environment variable
// (the latter is a side-effect of using Cobra/Viper and should NOT be documented because it's
// confusing).
// - The path to the kubeconfig file in the provided asset directory if assetDir is not an empty
// string.
// - The value provided via the KUBECONFIG environment variable.
// - ~/.kube/config (the default path kubectl uses).
func kubeconfigPath(assetDir string) string {
	var assetPath string
	if assetDir != "" {
		assetPath = filepath.Join(assetDir, "cluster-assets", "auth", "kubeconfig")
	}

	paths := []string{
		viper.GetString(kubeconfigFlag),
		assetPath,
		os.Getenv(kubeconfigEnvVariable),
		defaultKubeconfigPath,
	}

	for _, p := range paths {
		if p != "" {
			return p
		}
	}

	return ""
}

// askForConfirmation asks the user to confirm an action.
// It prints the message and then asks the user to type "yes" or "no".
// If the user types "yes" the function returns true, otherwise it returns
// false.
func askForConfirmation(message string) bool {
	var input string
	fmt.Printf("%s [type \"yes\" to continue]: ", message)
	fmt.Scanln(&input)
	return input == "yes"
}
