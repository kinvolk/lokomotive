// +build aws packet
// +build e2e

package components

import (
	"testing"

	"github.com/hashicorp/hcl/v2"

	"github.com/kinvolk/lokoctl/pkg/components"
	_ "github.com/kinvolk/lokoctl/pkg/components/flatcar-linux-update-operator"
	"github.com/kinvolk/lokoctl/pkg/components/util"
	testutil "github.com/kinvolk/lokoctl/test/components/util"
)

func TestInstallIdempotent(t *testing.T) {
	configHCL := `
component "flatcar-linux-update-operator" {}
  `

	n := "flatcar-linux-update-operator"

	c, err := components.Get(n)
	if err != nil {
		t.Fatalf("failed getting component: %v", err)
	}

	body, diagnostics := util.GetComponentBody(configHCL, n)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}

	diagnostics = c.LoadConfig(body, &hcl.EvalContext{})
	if diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}

	k := testutil.KubeconfigPath(t)
	if err := util.InstallAsRelease(n, c, k); err != nil {
		t.Fatalf("Installing component as relase should succeed, got: %v", err)
	}

	if err := util.InstallAsRelease(n, c, k); err != nil {
		t.Fatalf("Installing component twice as release should succeed, got: %v", err)
	}
}
