//+build e2e

package terraform

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func executor(t *testing.T) *Executor {
	tmpDir, err := ioutil.TempDir("", "lokoctl-tests-")
	if err != nil {
		t.Fatalf("Creating tmp dir should succeed, got: %v", err)
	}

	defer os.RemoveAll(tmpDir)

	conf := Config{
		Verbose:    false,
		WorkingDir: tmpDir,
	}

	ex, err := NewExecutor(conf)
	if err != nil {
		t.Fatalf("Creating new executor should succeed, got: %v", err)
	}

	return ex
}

func TestExecuteCheckErrors(t *testing.T) {
	ex := executor(t)

	if err := ex.Apply(); err == nil {
		t.Fatalf("Applying on empty directory should fail")
	}
}

func TestOutputIncludeKeyInError(t *testing.T) {
	ex := executor(t)

	k := "foo"
	o := ""

	err := ex.Output(k, &o)
	if err == nil {
		t.Fatalf("Output should fail on non existing installation")
	}

	if !strings.Contains(err.Error(), k) {
		t.Fatalf("Error message should contain key, got: %v", err)
	}
}
