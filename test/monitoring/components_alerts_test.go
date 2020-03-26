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

// +build aws packet aks
// +build poste2e

package monitoring

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

const (
	retryInterval  = time.Second * 5
	timeout        = time.Minute * 9
	contextTimeout = 10
)

type alertTestCase struct {
	ComponentName string
	RuleGroup     string
	platforms     []testutil.Platform
	Alerts        []string
}

//nolint:funlen
func testComponentAlerts(t *testing.T, v1api v1.API) {
	alertTestCases := []alertTestCase{
		{
			ComponentName: "metallb",
			RuleGroup:     "metallb-rules",
			platforms:     []testutil.Platform{testutil.PlatformPacket},
			Alerts: []string{
				"MetalLBNoBGPSession", "MetalLBConfigStale", "MetalLBControllerPodsAvailability",
				"MetalLBSpeakerPodsAvailability",
			},
		},
	}

	for _, tc := range alertTestCases {
		tc := tc
		t.Run(tc.ComponentName, func(t *testing.T) {
			t.Parallel()

			if !testutil.IsPlatformSupported(t, tc.platforms) {
				t.Skip()
			}

			if err := wait.PollImmediate(
				retryInterval, timeout, getComponentAlertRetryFunc(t, v1api, tc),
			); err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
}

func getComponentAlertRetryFunc(t *testing.T, v1api v1.API, tc alertTestCase) wait.ConditionFunc {
	return func() (done bool, err error) {
		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout*time.Second)
		defer cancel()

		result, err := v1api.Rules(ctx)
		if err != nil {
			return false, fmt.Errorf("error listing rules: %v", err)
		}

		// This map will store information from cluster so that it is easier to search it against
		// the test cases.
		ruleGroups := make(map[string][]string, len(result.Groups))

		for _, ruleGroup := range result.Groups {
			rules := make([]string, 0)

			for _, rule := range ruleGroup.Rules {
				switch v := rule.(type) {
				case v1.AlertingRule:
					rules = append(rules, v.Name)
				default:
				}
			}

			ruleGroups[ruleGroup.Name] = rules
		}

		rules, ok := ruleGroups[tc.RuleGroup]
		if !ok {
			// We don't return error here and just log it here because there is a
			// possibility that the prometheus has not reconciled and we need to just return
			// false i.e. not done and try again.
			t.Logf("error: RuleGroup %q not found. Retrying...", tc.RuleGroup)
			return false, nil
		}

		if !reflect.DeepEqual(rules, tc.Alerts) {
			return false, fmt.Errorf("Rules don't match. Expected: %#v and \ngot %#v", tc.Alerts, rules)
		}

		return true, nil
	}
}
