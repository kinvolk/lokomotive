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

	"github.com/kinvolk/lokomotive/pkg/terraform"
	"github.com/pkg/errors"
)

const (
	// Cloudflare represents DNS managed in Cloudflare.
	Cloudflare = "cloudflare"
	// Manual represents a manual DNS configuration.
	Manual = "manual"
	// Route53 represents DNS managed in Route 53.
	Route53 = "route53"
)

// Config represents a Lokomotive DNS configuration.
type Config struct {
	Provider string `hcl:"provider"`
	Zone     string `hcl:"zone"`
}

type dnsEntry struct {
	Name      string   `json:"name"`
	TTL       int      `json:"ttl"`
	EntryType string   `json:"type"`
	Records   []string `json:"records"`
}

// Validate ensures the specified DNS provider is valid.
func (c *Config) Validate() error {
	switch c.Provider {
	case Manual:
		return nil
	case Route53:
		return nil
	case Cloudflare:
		return nil
	}

	return fmt.Errorf("invalid DNS provider %q", c.Provider)
}

// ManualConfigPrompt returns a callback function which prompts the user to configure DNS entries
// manually and verifies the entries were created successfully.
func ManualConfigPrompt(c *Config) func(*terraform.Executor) error {
	return func(ex *terraform.Executor) error {
		dnsEntries, err := readDNSEntries(ex)
		if err != nil {
			return err
		}

		fmt.Printf("Please configure the following DNS entries at the DNS provider which hosts %q:\n", c.Zone)
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
		fmt.Printf("TTL: %d\n", entry.TTL)
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
