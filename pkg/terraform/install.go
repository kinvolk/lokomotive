package terraform

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hpcloud/tail"
	"github.com/pkg/errors"
)

func executeTerraform(ex *Executor, args ...string) error {
	pid, done, err := ex.Execute(args...)
	if err != nil {
		return errors.Wrapf(err, "failed to run 'terraform %s'", strings.Join(args, " "))
	}

	pathToFile := filepath.Join(ex.WorkingDirectory(), "logs", fmt.Sprintf("%d%s", pid, ".log"))
	t, err := tail.TailFile(pathToFile, tail.Config{Follow: true})
	if err != nil {
		return err
	}

	go func() {
		for line := range t.Lines {
			fmt.Println(line.Text)
		}
	}()

	<-done
	return t.Stop()
}

// InitAndApply create a new terraform executor for the given path,
// initializes it and applies the found modules.
func InitAndApply(exPath string) error {
	ex, err := NewExecutor(exPath)
	if err != nil {
		return errors.Wrap(err, "failed to create terraform executor")
	}

	if err := executeTerraform(ex, "init"); err != nil {
		return err
	}

	return executeTerraform(ex, "apply", "-auto-approve")
}
