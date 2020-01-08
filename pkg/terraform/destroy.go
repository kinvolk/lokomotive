package terraform

import (
	"github.com/pkg/errors"
)

// Destroy creates a new Terraform executor for the given path
// and executes `terraform destroy`.
func Destroy(exPath string) error {
	ex, err := NewExecutor(exPath)
	if err != nil {
		return errors.Wrap(err, "failed to create terraform executor")
	}

	return ExecuteTerraform(ex, "destroy", "-auto-approve")
}
