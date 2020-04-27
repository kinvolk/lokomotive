package aws

import (
	"testing"
)

func TestConfigNewconfig(t *testing.T) {
	c := NewConfig()

	if c.ControllerType != "t3.small" {
		t.Fatalf("Expected default controller Type to be `t3.small`, got: %s", c.ControllerType)
	}

	if c.ClusterDomainSuffix != "cluster.local" {
		t.Fatalf("Expected default ClusterDomainSufix as `cluster.local`, got: %s", c.ClusterDomainSuffix)
	}

	if c.CertsValidityPeriodHours != 8760 {
		t.Fatalf("Expected default CertsValidityPeriodHours as `8760`, got: %d", c.CertsValidityPeriodHours)
	}

	if !c.EnableAggregation {
		t.Fatalf("Expected default EnableAggreation to be true, got: %t", c.EnableAggregation)
	}

	if c.ControllerCount != 1 {
		t.Fatalf("Expected default ControllerCount to be 1, got: %d", c.ControllerCount)
	}

	if c.OSChannel != "stable" {
		t.Fatalf("Expected default channel as `stable`, got: %s", c.OSChannel)
	}

	if c.OSVersion != "current" {
		t.Fatalf("Expected default version as `current`, got: %s", c.OSVersion)
	}

	if c.NetworkMTU != 1480 {
		t.Fatalf("Expected default NetworkMTU as `1480`, got: %d", c.NetworkMTU)
	}

	if c.PodCIDR != "10.2.0.0/16" {
		t.Fatalf("Expected default PodCIDR as `10.2.0.0/16`, got: %s", c.PodCIDR)
	}

	if c.ServiceCIDR != "10.3.0.0/16" {
		t.Fatalf("Expected default ServiceCIDR as `10.3.0.0/16`, got: %s", c.ServiceCIDR)
	}

	if c.EnableReporting {
		t.Fatalf("Expected default EnableReporting as false, got: %t", c.EnableReporting)
	}
}

func TestConfigValidateSuccess(t *testing.T) {
	a := &config{
		AssetDir:                 "test-asset-dir",
		DNSZone:                  "test-zone",
		DNSZoneID:                "test-zone-id",
		Region:                   "eu-central-1",
		OSVersion:                "current",
		OSChannel:                "stable",
		ClusterName:              "test",
		SSHPubKeys:               []string{"ssh-key"},
		PodCIDR:                  "10.2.0.0/16",
		ServiceCIDR:              "10.3.0.0/16",
		HostCIDR:                 "10.0.0.0/16",
		ControllerType:           "t3.small",
		ClusterDomainSuffix:      "test.local",
		ControllerCount:          3,
		CertsValidityPeriodHours: 1000,
		NetworkMTU:               2000,
		DiskType:                 "eb2",
		DiskSize:                 40,
		WorkerPools: []workerPool{
			{
				Name:  "pool",
				Count: 3,
			},
		},
	}

	diags := a.Validate()
	if diags.HasErrors() {
		for _, diag := range diags {
			t.Error(diag)
		}

		t.Fatalf("Expected no errors in validating configuration, got: %s", diags.Error())
	}
}

func TestConfigValidateFail(t *testing.T) {
	p := &config{
		AssetDir:                 "",
		Region:                   "",
		OSVersion:                "current",
		OSChannel:                "asd",
		ClusterName:              "test",
		CertsValidityPeriodHours: 1000,
		NetworkMTU:               2000,
		SSHPubKeys:               []string{"ssh-key"},
		PodCIDR:                  "C.2.0.0/16",
		ServiceCIDR:              "10.X.0.0/16",
		HostCIDR:                 "10.0.0.0/16",
		ControllerType:           "",
		ClusterDomainSuffix:      "test.local",
		ControllerCount:          1,
		WorkerPools: []workerPool{
			{
				Name:  "pool",
				Count: 3,
			},
		},
	}

	diags := p.Validate()
	if !diags.HasErrors() {
		t.Fatalf("Expected errors in validating configuration")
	}
}

func TestRenderSuccess(t *testing.T) {
	p := &config{
		AssetDir:                 "test-asset-dir",
		DNSZone:                  "test-zone",
		DNSZoneID:                "test-zone-id",
		Region:                   "eu-central-1",
		OSVersion:                "current",
		OSChannel:                "stable",
		ClusterName:              "test",
		CertsValidityPeriodHours: 1000,
		NetworkMTU:               2000,
		SSHPubKeys:               []string{"ssh-key"},
		PodCIDR:                  "10.2.0.0/16",
		ServiceCIDR:              "10.3.0.0/16",
		HostCIDR:                 "10.0.0.0/16",
		ControllerType:           "baremetal_0",
		ClusterDomainSuffix:      "test.local",
		ControllerCount:          3,
		WorkerPools: []workerPool{
			{
				Name:  "pool",
				Count: 3,
			},
		},
	}

	renderedTemplate, err := p.Render()
	if err != nil {
		t.Fatalf("Expected render to succeed, got: %v", err)
	}

	if renderedTemplate == "" {
		t.Fatalf("Expected rendered string to be non-empty")
	}
}
