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

	"github.com/hashicorp/hcl/v2"
	"github.com/kinvolk/lokomotive/pkg/lokomotive/config"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
	"github.com/sirupsen/logrus"
)

// lokomotive manages the Lokomotive cluster related operations such as Apply,
// Destroy ,Health etc.
type lokomotive struct {
	ContextLogger *logrus.Entry
	Platform      platform.Platform
	Config        *config.LokomotiveConfig
	Executor      *terraform.Executor
}

// NewLokomotive returns the an new lokomotive Instance
func NewLokomotive(ctxLogger *logrus.Entry, cfg *config.LokomotiveConfig, options *Options) (Manager, hcl.Diagnostics) {
	// Initialize Terraform Executor
	ex, err := terraform.InitializeExecutor(cfg.Platform.GetAssetDir(), options.Verbose)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to initialize Terraform executor: %v", err),
		}

		return nil, hcl.Diagnostics{diag}
	}

	return &lokomotive{
		ContextLogger: ctxLogger,
		Config:        cfg,
		Platform:      cfg.Platform,
		Executor:      ex,
	}, hcl.Diagnostics{}
}

func (l *lokomotive) Apply(*Options) {

}

func (l *lokomotive) Destroy(*Options) {
}

func (l *lokomotive) ApplyComponents([]string) {

}

func (l *lokomotive) RenderComponents([]string) {

}

func (l *lokomotive) Health() {

}
