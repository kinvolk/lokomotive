package aks_test

// TODO: Migrate tests.

// // Meta()
// func TestMeta(t *testing.T) {
// 	assetDir := "foo"

// 	moreWorkers := 3

// 	c := &config{
// 		AssetDir: assetDir,
// 		WorkerPools: []workerPool{
// 			{
// 				Count: testWorkerCount,
// 			},
// 			{
// 				Count: moreWorkers,
// 			},
// 		},
// 	}

// 	expectedNodes := 4
// 	if e := c.Meta().ExpectedNodes; e != expectedNodes {
// 		t.Errorf("Meta should count workers from all pools. Expected %d, got %d", expectedNodes, e)
// 	}

// 	if a := c.Meta().AssetDir; a != assetDir {
// 		t.Errorf("Meta should return configured asset dir. Expected %q, got %q", assetDir, a)
// 	}
// }

// // checkWorkerPoolNamesUnique()
// func TestCheckWorkerPoolNamesUniqueDuplicated(t *testing.T) {
// 	c := &config{
// 		WorkerPools: []workerPool{
// 			{
// 				Name: "foo",
// 			},
// 			{
// 				Name: "foo",
// 			},
// 		},
// 	}

// 	if d := c.checkWorkerPoolNamesUnique(); !d.HasErrors() {
// 		t.Fatalf("should return error when worker pools are duplicated")
// 	}
// }

// func TestCheckWorkerPoolNamesUnique(t *testing.T) {
// 	c := &config{
// 		WorkerPools: []workerPool{
// 			{
// 				Name: "foo",
// 			},
// 			{
// 				Name: "bar",
// 			},
// 		},
// 	}

// 	if d := c.checkWorkerPoolNamesUnique(); d.HasErrors() {
// 		t.Fatalf("should not return errors when pool names are unique, got: %v", d)
// 	}
// }

// // checkNotEmptyWorkers()
// func TestNotEmptyWorkersEmpty(t *testing.T) {
// 	c := &config{}

// 	if d := c.checkNotEmptyWorkers(); !d.HasErrors() {
// 		t.Fatalf("should return error when there is no worker pool defined")
// 	}
// }

// func TestNotEmptyWorkers(t *testing.T) {
// 	c := &config{
// 		WorkerPools: []workerPool{
// 			{
// 				Name: "foo",
// 			},
// 		},
// 	}

// 	if d := c.checkNotEmptyWorkers(); d.HasErrors() {
// 		t.Fatalf("should not return errors when worker pool is not empty, got: %v", d)
// 	}
// }

// // checkConfiguration()
// func TestCheckWorkerPoolNamesUniqueTest(t *testing.T) {
// 	c := &config{
// 		WorkerPools: []workerPool{
// 			{
// 				Name: "foo",
// 			},
// 			{
// 				Name: "bar",
// 			},
// 		},
// 	}

// 	if d := c.checkWorkerPoolNamesUnique(); d.HasErrors() {
// 		t.Fatalf("should not return errors when pool names are unique, got: %v", d)
// 	}
// }

// // checkCredentials()
// func TestCheckCredentialsAppNameAndClientID(t *testing.T) {
// 	c := &config{
// 		ApplicationName: "foo",
// 		ClientID:        "foo",
// 	}

// 	if d := c.checkCredentials(); !d.HasErrors() {
// 		t.Fatalf("should give error if both ApplicationName and ClientID fields are defined")
// 	}
// }

// func TestCheckCredentialsAppNameAndClientSecret(t *testing.T) {
// 	c := &config{
// 		ApplicationName: "foo",
// 		ClientSecret:    "foo",
// 	}

// 	if d := c.checkCredentials(); !d.HasErrors() {
// 		t.Fatalf("should give error if both ApplicationName and ClientID fields are defined")
// 	}
// }

// func TestCheckCredentialsAppNameClientIDAndClientSecret(t *testing.T) {
// 	c := &config{
// 		ApplicationName: "foo",
// 		ClientID:        "foo",
// 		ClientSecret:    "foo",
// 	}

// 	expectedErrorCount := 2

// 	if d := c.checkCredentials(); len(d) != expectedErrorCount {
// 		t.Fatalf("should give errors for both conflicting ClientID and ClientSecret, got %v", d)
// 	}
// }

// func TestCheckCredentialsRequireSome(t *testing.T) {
// 	c := &config{}

// 	if d := c.checkCredentials(); !d.HasErrors() {
// 		t.Fatalf("should give error if both ApplicationName and ClientID fields are empty")
// 	}
// }

// func TestCheckCredentialsRequireClientIDWithClientSecret(t *testing.T) {
// 	c := &config{
// 		ClientSecret: "foo",
// 	}

// 	if d := c.checkCredentials(); !d.HasErrors() {
// 		t.Fatalf("should give error if ClientSecret is defined and ClientID is empty")
// 	}
// }

// func TestCheckCredentialsReadClientSecretFromEnvironment(t *testing.T) {
// 	if err := os.Setenv(clientSecretEnv, "1"); err != nil {
// 		t.Fatalf("failed to set environment variable %q: %v", clientSecretEnv, err)
// 	}

// 	defer func() {
// 		if err := os.Setenv(clientSecretEnv, ""); err != nil {
// 			t.Logf("failed unsetting environment variable %q: %v", clientSecretEnv, err)
// 		}
// 	}()

// 	c := &config{
// 		ClientID: "foo",
// 	}

// 	if d := c.checkCredentials(); d.HasErrors() {
// 		t.Fatalf("should read client secret from environment")
// 	}
// }

// // LoadConfig()
// func loadConfigFromString(t *testing.T, c string) hcl.Diagnostics {
// 	p := hclparse.NewParser()

// 	f, d := p.ParseHCL([]byte(c), "x.lokocfg")
// 	if d.HasErrors() {
// 		t.Fatalf("parsing HCL should succeed, got: %v", d)
// 	}

// 	configBody := hcl.MergeFiles([]*hcl.File{f})

// 	var rootConfig lokoconfig.RootConfig

// 	if d := gohcl.DecodeBody(configBody, nil, &rootConfig); d.HasErrors() {
// 		t.Fatalf("decoding root config should succeed, got: %v", d)
// 	}

// 	cc := &config{}

// 	return cc.LoadConfig(&rootConfig.Cluster.Config, &hcl.EvalContext{})
// }

// func TestLoadConfig(t *testing.T) {
// 	c := `
// cluster "aks" {
//   asset_dir           = "/fooo"
//   client_id           = "bar"
//   client_secret       = "foo"
//   cluster_name        = "mycluster"
//   resource_group_name = "test"
//   subscription_id     = "foo"
//   tenant_id           = "bar"

//   worker_pool "foo" {
//     count   = 1
//     vm_size = "foo"
//   }
// }
// `
// 	if d := loadConfigFromString(t, c); d.HasErrors() {
// 		t.Fatalf("valid config should not return error, got: %v", d)
// 	}
// }

// func TestLoadConfigEmpty(t *testing.T) {
// 	c := &config{}

// 	if d := c.LoadConfig(nil, &hcl.EvalContext{}); !d.HasErrors() {
// 		t.Fatalf("empty config should return error, got: %v", d)
// 	}
// }

// func TestLoadConfigBadHCL(t *testing.T) {
// 	c := `
// cluster "aks" {
//   not_defined_field = "doh"
// }
// `

// 	if d := loadConfigFromString(t, c); !d.HasErrors() {
// 		t.Fatalf("invalid HCL should return errors")
// 	}
// }
