//+build e2e

package terraform

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestExecuteCheckErrors(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "lokoctl-tests-")
	if err != nil {
		t.Fatalf("Creating tmp dir should succeed, got: %v", err)
	}

	defer os.RemoveAll(tmpDir)

	conf := Config{
		Quiet:      true,
		WorkingDir: tmpDir,
	}

	ex, err := NewExecutor(conf)
	if err != nil {
		t.Fatalf("Creating new executor should succeed, got: %v", err)
	}

	if err := ex.Execute("apply"); err == nil {
		t.Fatalf("Applying on empty directory should fail")
	}
}
