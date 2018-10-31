package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

func isKubeconfigSet(cmd *cobra.Command, args []string) error {
	if kubeconfig == "" {
		return errors.New(`required flag "kubeconfig" not set`)
	}

	return nil
}
