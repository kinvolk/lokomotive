package kubernetes_test

import (
	"time"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

const (
	namespace = "kube-system"
)

func components() map[string]time.Duration {
	return map[string]time.Duration{
		// If we kill active kube-controller-manager, it takes time for new one to kick in
		// so we need to wait longer.
		"kube-controller-manager": 1 * time.Minute,
		"kube-scheduler":          testutil.RetryInterval,
		"kube-apiserver":          testutil.RetryInterval,
		"coredns":                 testutil.RetryInterval,
	}
}
