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

// +build packet
// +build poste2e

package packet_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	contextTimeout = 10
)

type componentTestCase struct {
	componentName string
	platforms     []testutil.Platform
	namespace     string
}

func TestIngressHost(t *testing.T) {
	client := testutil.CreateKubeClient(t)
	componentTestCases := []componentTestCase{
		{
			componentName: "dex",
			platforms:     []testutil.Platform{testutil.PlatformPacket},
			namespace:     "dex",
		},
		{
			componentName: "gangway",
			platforms:     []testutil.Platform{testutil.PlatformPacket},
			namespace:     "gangway",
		},
	}

	for _, tc := range componentTestCases {
		tc := tc
		t.Run(tc.componentName, func(t *testing.T) {
			t.Parallel()

			if !testutil.IsPlatformSupported(t, tc.platforms) {
				t.Skip()
			}

			if err := wait.PollImmediate(
				testutil.RetryInterval, testutil.Timeout, checkIngressHost(client, tc),
			); err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
}

func checkIngressHost(client kubernetes.Interface, tc componentTestCase) wait.ConditionFunc {
	return func() (done bool, err error) {
		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout*time.Second)
		defer cancel()

		ing, err := client.ExtensionsV1beta1().Ingresses(tc.namespace).Get(ctx, tc.componentName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("error getting ingress: %v", err)
		}

		// DNS records are added using external-dns
		// which is configured to use Route53

		ingIP := ing.Status.LoadBalancer.Ingress[0].IP
		ingHost := ing.Spec.Rules[0].Host

		addr, err := net.LookupIP(ingHost)

		if err != nil {
			return false, fmt.Errorf("unknown host: %v", err)
		}

		for _, v := range addr {
			if v.String() == ingIP {
				return true, nil
			}
		}

		return false, nil
	}
}
