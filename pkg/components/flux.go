// Copyright 2021 The Lokomotive Authors
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

// Package components hosts generic stuff needed by the individual components.
package components

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// FluxInstallTimeout is used in creation of HelmRelease object by individual components.
	FluxInstallTimeout = metav1.Duration{Duration: time.Minute * 10} //nolint:gomnd

	// FluxInstallInterval is used in creation of HelmRelease object by individual components.
	FluxInstallInterval = metav1.Duration{Duration: time.Minute}

	// ComponentsPath is the relative path of the component assets in the Lokomotive project.
	ComponentsPath = "./assets/charts/components/"
)
