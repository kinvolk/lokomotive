// Copyright 2020 The Lokomotive Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dns

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"sort"

	"github.com/kinvolk/lokoctl/pkg/terraform"
	"github.com/pkg/errors"
)

type DNSProvider int

const (
	DNSNone DNSProvider = iota
	DNSManual
	DNSRoute53
)

type route53Provider struct {
	ZoneID       string `hcl:"zone_id,optional"`
	AWSCredsPath string `hcl:"aws_creds_path,optional"`
}

type dnsProvider struct {
	Route53 *route53Provider `hcl:"route53,block"`
	Manual  *manualProvider  `hcl:"manual,block"`
}

type manualProvider struct{}

type Config struct {
	Zone     string      `hcl:"zone"`
	Provider dnsProvider `hcl:"provider,block"`
}

type dnsEntry struct {
	Name      string   `json:"name"`
	Ttl       int      `json:"ttl"`
	EntryType string   `json:"type"`
	Records   []string `json:"records"`
}

// ParseDNS checks that the DNS provider configuration is correct and returns
// the configured provider.
func ParseDNS(config *Config) (DNSProvider, error) {
	// Check that only one provider is specified.
	if config.Provider.Manual != nil && config.Provider.Route53 != nil {
		return DNSNone, fmt.Errorf("multiple DNS providers specified")
	}

	if config.Provider.Manual != nil {
		return DNSManual, nil
	}

	if config.Provider.Route53 != nil {
		return DNSRoute53, nil
	}

	return DNSNone, fmt.Errorf("no DNS provider specified")
}

// AskToConfigure reads the required DNS entries from a Terraform output,
// asks the user to configure them and checks if the configuration is correct.
func AskToConfigure(ex *terraform.Executor, cfg *Config) error {
	dnsEntries, err := readDNSEntries(ex)
	if err != nil {
		return err
	}

	fmt.Printf("Please configure the following DNS entries at the DNS provider which hosts %q:\n", cfg.Zone)
	prettyPrintDNSEntries(dnsEntries)

	for {
		fmt.Printf("Press Enter to check the entries or type \"skip\" to continue the installation: ")

		var input string
		fmt.Scanln(&input)

		if input == "skip" {
			break
		} else if input != "" {
			continue
		}

		if checkDNSEntries(dnsEntries) {
			break
		}

		fmt.Println("Entries are not correctly configured, please verify.")
	}

	return nil
}

func readDNSEntries(ex *terraform.Executor) ([]dnsEntry, error) {
	output, err := ex.ExecuteSync("output", "-json", "dns_entries")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get DNS entries")
	}

	var entries []dnsEntry

	if err := json.Unmarshal(output, &entries); err != nil {
		return nil, errors.Wrap(err, "failed to parse DNS entries file")
	}

	return entries, nil
}

func prettyPrintDNSEntries(entries []dnsEntry) {
	fmt.Println("------------------------------------------------------------------------")

	for _, entry := range entries {
		fmt.Printf("Name: %s\n", entry.Name)
		fmt.Printf("Type: %s\n", entry.EntryType)
		fmt.Printf("Ttl: %d\n", entry.Ttl)
		fmt.Printf("Records:\n")
		for _, record := range entry.Records {
			fmt.Printf("- %s\n", record)
		}
		fmt.Println("------------------------------------------------------------------------")
	}
}

func checkDNSEntries(entries []dnsEntry) bool {
	for _, entry := range entries {
		ips, err := net.LookupIP(entry.Name)
		if err != nil {
			return false
		}

		var ipsString []string
		for _, ip := range ips {
			ipsString = append(ipsString, ip.String())
		}

		sort.Strings(ipsString)
		sort.Strings(entry.Records)
		if !reflect.DeepEqual(ipsString, entry.Records) {
			return false
		}
	}

	return true
}
