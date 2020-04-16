package config

import (
	"testing"
)

func TestNetworkConfigDefaultConfig(t *testing.T) {
	n := DefaultNetworkConfig()

	if n.NetworkMTU != defaultNetworkMTU {
		t.Fatalf("Expected default NetworkMTU as `1480`, got: %d", n.NetworkMTU)
	}

	if n.PodCIDR != "10.2.0.0/16" {
		t.Fatalf("Expected default PodCIDR as `10.2.0.0/16`, got: %s", n.PodCIDR)
	}

	if n.ServiceCIDR != "10.3.0.0/16" {
		t.Fatalf("Expected default ServiceCIDR as `10.3.0.0/16`, got: %s", n.ServiceCIDR)
	}

	if n.EnableReporting {
		t.Fatalf("Expected default EnableReporting as false, got: %t", n.EnableReporting)
	}
}

func TestNetworkConfigValidateSuccess(t *testing.T) {
	n := &NetworkConfig{
		//nolint:gomnd
		NetworkMTU:      2000,
		PodCIDR:         "10.11.0.0/16",
		ServiceCIDR:     "10.12.0.0/16",
		EnableReporting: true,
	}

	diags := n.Validate()
	if diags.HasErrors() {
		t.Fatalf("Expected no errors in validating configuration")
	}
}

func TestNetworkConfigValidateFail(t *testing.T) {
	n := &NetworkConfig{
		NetworkMTU:  0,
		PodCIDR:     "",
		ServiceCIDR: "0.0.0.0/16",
	}

	diags := n.Validate()
	if !diags.HasErrors() {
		t.Fatalf("Expected errors in validating configuration, got: %s", diags.Error())
	}

	//nolint:gomnd
	if len(diags) == 3 {
		t.Fatalf("Expected three errors - NetworkMTU more than 0, PodCIDR and ServiceCIDR invalid, got: %d", len(diags))
	}
}
