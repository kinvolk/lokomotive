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

// +build aws aws_edge packet aks
// +build poste2e

package monitoring

import (
	"context"
	"fmt"
	"os"
	"testing"
	"text/tabwriter"
	"time"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func testScrapeTargetRechability(t *testing.T, v1api v1.API) {
	var w *tabwriter.Writer

	if err := wait.PollImmediate(
		testutil.RetryInterval, testutil.TimeoutSlow, getScrapeTargetRetryFunc(t, v1api, w),
	); err != nil {
		t.Errorf("%v", err)

		if w == nil {
			return
		}

		// Finally print the table of all the targets that are down.
		if err := w.Flush(); err != nil {
			t.Errorf("error printing the unreachable targets: %v", err)
		}
	}
}

func getScrapeTargetRetryFunc(t *testing.T, v1api v1.API, w *tabwriter.Writer) wait.ConditionFunc {
	return func() (done bool, err error) {
		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout*time.Second)
		defer cancel()

		targets, err := v1api.Targets(ctx)
		if err != nil {
			return false, fmt.Errorf("error listing targets from prometheus: %v", err)
		}

		// Initialize the tabwriter to print the output in tabular format.
		w = new(tabwriter.Writer)
		w.Init(os.Stdout, 16, 8, 2, '\t', 0)
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "Service\tHealth\n")
		fmt.Fprintf(w, "-------\t------\n")

		// Boolean used to identify if tests failed.
		var testsFailed bool

		for _, target := range targets.Active {
			if target.Health == v1.HealthGood {
				continue
			}

			// This variable marks that the test has failed but we don't return from here because we
			// need the list of all the targets that are not in UP state.
			testsFailed = true

			fmt.Fprintf(w, "%s/%s\t%s\n",
				target.Labels["namespace"], target.Labels["service"], target.Health)
		}

		fmt.Fprintf(w, "\n")

		if testsFailed {
			t.Logf("Some prometheus scrape targets are down. Retrying ...")
			return false, nil
		}

		return true, nil
	}
}
