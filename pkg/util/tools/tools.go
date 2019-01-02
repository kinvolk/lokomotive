package tools

import (
	"os"
	"os/exec"

	k8serrs "k8s.io/apimachinery/pkg/util/errors"
)

func isBinaryInPath(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return err
	}
	return nil
}

// InstallerBinaries is used from sub-command `install` to check if all the
// needed binaries are in required place
func InstallerBinaries() error {
	var errs []error

	if err := isBinaryInPath("terraform"); err != nil {
		errs = append(errs, err)
	}

	// see if the terraform plugin is in path
	plugin := "$HOME/.terraform.d/plugins/terraform-provider-ct"
	if _, err := os.Stat(os.ExpandEnv(plugin)); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return k8serrs.NewAggregate(errs)
	}
	return nil
}
