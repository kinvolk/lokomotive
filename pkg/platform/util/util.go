package util

import (
	"fmt"
	"net"

	"github.com/hashicorp/hcl/v2"
	"github.com/kinvolk/lokomotive/pkg/version"
)

func AppendTags(tags *map[string]string) {
	if tags == nil {
		return
	}
	if *tags == nil {
		*tags = make(map[string]string)
	}
	if version.Version != "" {
		(*tags)["lokoctl-version"] = version.Version
	}
}

func IsFlatcarChannelSupported(c string) bool {
	supported := false

	supportedChannels := []string{"stable", "alpha", "beta", "edge"}
	for _, channel := range supportedChannels {
		if c == channel {
			supported = true
		}
	}

	return supported
}

func IsValidCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)

	return err
}

func IsValidOSArch(a string) bool {
	valid := false

	archs := []string{"amd64", "arm64"}
	for _, arch := range archs {
		if a == arch {
			valid = true
		}
	}

	return valid
}

func CheckIsEmptyField(data, nameOfField string) hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if data == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("'%s' cannot be empty", nameOfField),
		})
	}

	return diagnostics
}
