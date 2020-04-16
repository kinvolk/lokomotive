package config

import (
	"testing"
)

func TestDefaultControllerConfig(t *testing.T) {
	c := DefaultControllerConfig()

	if c.Count != defaultControllerCount {
		t.Fatalf("Expected default value of count as 1, got: %d", c.Count)
	}
}

func TestControllerConfigValidateSuccess(t *testing.T) {
	c := &ControllerConfig{
		//nolint:gomnd
		Count:      2,
		SSHPubKeys: []string{"test-ssh-key"},
	}

	diags := c.Validate()
	if diags.HasErrors() {
		t.Fatalf("Expected no errors, got: %s", diags.Error())
	}
}

func TestControllerConfigValidateFail(t *testing.T) {
	c := &ControllerConfig{
		Count:      0,
		SSHPubKeys: []string{},
	}

	diags := c.Validate()
	if !diags.HasErrors() {
		t.Fatalf("Expected errors, found none")
	}
	//nolint:gomnd
	if len(diags) != 2 {
		t.Fatalf("Expected two errors, one of count and another for empty ssh-keys, got: %d", len(diags))
	}
}
