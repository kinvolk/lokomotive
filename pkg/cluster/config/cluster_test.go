package config

import (
	"testing"
)

func TestClusterDefaultConfig(t *testing.T) {
	c := DefaultClusterConfig()

	if c.ClusterDomainSuffix != "cluster.local" {
		t.Fatalf("Expected default ClusterDomainSufix as `cluster.local`, got: %s", c.ClusterDomainSuffix)
	}

	if c.CertsValidityPeriodHours != defaultCertsValidityPeriodHours {
		t.Fatalf("Expected default CertsValidityPeriodHours as `8760`, got: %d", c.CertsValidityPeriodHours)
	}

	if !c.EnableAggregation {
		t.Fatalf("Expected default EnableAggreation to be true, got: %t", c.EnableAggregation)
	}
}

func TestClusterConfigValidateSuccess(t *testing.T) {
	c := ClusterConfig{
		//nolint:gomnd
		CertsValidityPeriodHours: 1000,
		ClusterDomainSuffix:      "test.local",
	}

	diags := c.Validate()
	if diags.HasErrors() {
		t.Fatalf("Expected no errors in validating configuration, got: %s", diags.Error())
	}
}

func TestClusterConfigValidateFail(t *testing.T) {
	m := ClusterConfig{
		ClusterDomainSuffix:      "",
		CertsValidityPeriodHours: 0,
	}

	diags := m.Validate()
	if !diags.HasErrors() {
		t.Fatalf("Expected errors in validating configuration")
	}
	//nolint:gomnd
	if len(diags) != 2 {
		t.Fatalf("Expected four errors - empty ClusterDomainSuffix"+
			"and invalid CertsValidityPeriodHours, got: %d", len(diags))
	}
}
