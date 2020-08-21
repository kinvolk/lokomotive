package aks

import (
	"os"
	"testing"
)

const (
	testWorkerCount = 1
)

// TODO: Increase test coverage.

func TestRenderRootModule(t *testing.T) {
	c := &Config{
		WorkerPools: []workerPool{
			{
				Name:   "foo",
				VMSize: "bar",
				Count:  testWorkerCount,
			},
		},
	}

	_, err := renderRootModule(c)
	if err != nil {
		t.Fatalf("Rendering root module: %v", err)
	}
}

func TestCheckWorkerPoolNamesUniqueDuplicated(t *testing.T) {
	c := &Config{
		WorkerPools: []workerPool{
			{
				Name: "foo",
			},
			{
				Name: "foo",
			},
		},
	}

	if d := c.checkWorkerPoolNamesUnique(); !d.HasErrors() {
		t.Fatalf("should return error when worker pools are duplicated")
	}
}

func TestCheckWorkerPoolNamesUnique(t *testing.T) {
	c := &Config{
		WorkerPools: []workerPool{
			{
				Name: "foo",
			},
			{
				Name: "bar",
			},
		},
	}

	if d := c.checkWorkerPoolNamesUnique(); d.HasErrors() {
		t.Fatalf("should not return errors when pool names are unique, got: %v", d)
	}
}

func TestNotEmptyWorkersEmpty(t *testing.T) {
	c := &Config{}

	if d := c.checkNotEmptyWorkers(); !d.HasErrors() {
		t.Fatalf("should return error when there is no worker pool defined")
	}
}

func TestNotEmptyWorkers(t *testing.T) {
	c := &Config{
		WorkerPools: []workerPool{
			{
				Name: "foo",
			},
		},
	}

	if d := c.checkNotEmptyWorkers(); d.HasErrors() {
		t.Fatalf("should not return errors when worker pool is not empty, got: %v", d)
	}
}

func TestCheckWorkerPoolNamesUniqueTest(t *testing.T) {
	c := &Config{
		WorkerPools: []workerPool{
			{
				Name: "foo",
			},
			{
				Name: "bar",
			},
		},
	}

	if d := c.checkWorkerPoolNamesUnique(); d.HasErrors() {
		t.Fatalf("should not return errors when pool names are unique, got: %v", d)
	}
}

func TestCheckCredentialsAppNameAndClientID(t *testing.T) {
	c := &Config{
		ApplicationName: "foo",
		ClientID:        "foo",
	}

	if d := c.checkCredentials(); !d.HasErrors() {
		t.Fatalf("should give error if both ApplicationName and ClientID fields are defined")
	}
}

func TestCheckCredentialsAppNameAndClientSecret(t *testing.T) {
	c := &Config{
		ApplicationName: "foo",
		ClientSecret:    "foo",
	}

	if d := c.checkCredentials(); !d.HasErrors() {
		t.Fatalf("should give error if both ApplicationName and ClientID fields are defined")
	}
}

func TestCheckCredentialsAppNameClientIDAndClientSecret(t *testing.T) {
	c := &Config{
		ApplicationName: "foo",
		ClientID:        "foo",
		ClientSecret:    "foo",
	}

	expectedErrorCount := 2

	if d := c.checkCredentials(); len(d) != expectedErrorCount {
		t.Fatalf("should give errors for both conflicting ClientID and ClientSecret, got %v", d)
	}
}

func TestCheckCredentialsRequireSome(t *testing.T) {
	c := &Config{}

	if d := c.checkCredentials(); !d.HasErrors() {
		t.Fatalf("should give error if both ApplicationName and ClientID fields are empty")
	}
}

func TestCheckCredentialsRequireClientIDWithClientSecret(t *testing.T) {
	c := &Config{
		ClientSecret: "foo",
	}

	if d := c.checkCredentials(); !d.HasErrors() {
		t.Fatalf("should give error if ClientSecret is defined and ClientID is empty")
	}
}

func TestCheckCredentialsReadClientSecretFromEnvironment(t *testing.T) {
	if err := os.Setenv(clientSecretEnv, "1"); err != nil {
		t.Fatalf("failed to set environment variable %q: %v", clientSecretEnv, err)
	}

	defer func() {
		if err := os.Setenv(clientSecretEnv, ""); err != nil {
			t.Logf("failed unsetting environment variable %q: %v", clientSecretEnv, err)
		}
	}()

	c := &Config{
		ClientID: "foo",
	}

	if d := c.checkCredentials(); d.HasErrors() {
		t.Fatalf("should read client secret from environment")
	}
}
