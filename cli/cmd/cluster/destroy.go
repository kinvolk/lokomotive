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

package cluster

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// DestroyOptions controls Destroy() behavior.
type DestroyOptions struct {
	Confirm    bool
	Verbose    bool
	ConfigPath string
	ValuesPath string
}

// Destroy destroys cluster infrastructure.
func Destroy(contextLogger *log.Entry, options DestroyOptions) error {
	cc := clusterConfig{
		verbose:    options.Verbose,
		configPath: options.ConfigPath,
		valuesPath: options.ValuesPath,
	}

	c, err := cc.initialize(contextLogger)
	if err != nil {
		return fmt.Errorf("initializing: %w", err)
	}

	exists, err := clusterExists(c.terraformExecutor)
	if err != nil {
		return fmt.Errorf("checking if cluster exists: %w", err)
	}

	if !exists {
		contextLogger.Println("Cluster already destroyed, nothing to do")

		return nil
	}

	if !options.Confirm {
		confirmation := askForConfirmation("WARNING: This action cannot be undone. " +
			"Do you really want to destroy the cluster?")

		if !confirmation {
			contextLogger.Println("Cluster destroy canceled")

			return nil
		}
	}

	if err := c.platform.Destroy(&c.terraformExecutor); err != nil {
		return fmt.Errorf("destroying cluster: %v", err)
	}

	contextLogger.Println("Cluster destroyed successfully")
	contextLogger.Println("You can safely remove the assets directory now")

	return nil
}
