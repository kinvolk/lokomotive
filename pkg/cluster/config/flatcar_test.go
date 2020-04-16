package config

import (
	"testing"
)

func TestFlatcarConfigDefaultConfig(t *testing.T) {
	f := DefaultFlatcarConfig()

	if f.Channel != "stable" {
		t.Fatalf("Expected default channel as `stable`, got: %s", f.Channel)
	}

	if f.Version != "current" {
		t.Fatalf("Expected default version as `current`, got: %s", f.Version)
	}
}

func TestFlatcarConfigValidateSuccess(t *testing.T) {
	f := &FlatcarConfig{
		Channel: "alpha",
		Version: "2301.1.1",
	}

	diags := f.Validate()
	if diags.HasErrors() {
		t.Fatalf("Expected no errors in validation, got: %s", diags.Error())
	}
}

func TestFlatcarConfigValidateFail(t *testing.T) {
	f := &FlatcarConfig{
		Channel: "test",
		Version: "",
	}

	diags := f.Validate()
	if !diags.HasErrors() {
		t.Fatalf("Expected errors in validating configuration")
	}

	//nolint:gomnd
	if len(diags) != 2 {
		t.Fatalf("Expected two errors, unsupported channel and empty version, got: %d", len(diags))
	}
}
