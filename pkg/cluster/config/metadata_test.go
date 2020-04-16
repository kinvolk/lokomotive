package config

import (
	"testing"
)

func TestMetadataValidateSuccess(t *testing.T) {
	m := Metadata{
		ClusterName: "test-cluster",
		AssetDir:    "path/to/asset/dir",
	}

	diags := m.Validate()
	if diags.HasErrors() {
		t.Fatalf("Expected no errors in validating configuration, got: %s", diags.Error())
	}
}

func TestMetadataValidateFail(t *testing.T) {
	m := Metadata{
		ClusterName: "",
		AssetDir:    "",
	}

	diags := m.Validate()
	if !diags.HasErrors() {
		t.Fatalf("Expected errors in validating configuration")
	}

	//nolint:gomnd
	if len(diags) != 2 {
		t.Fatalf("Expected four errors - empty ClusterName, AssetDir, got: %d", len(diags))
	}
}
