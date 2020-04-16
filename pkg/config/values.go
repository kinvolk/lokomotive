// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// LoadValues reads the list of hcl.File and parses it as a
// "values file" (flat key.value HCL config) for later use in the
// `EvalContext`.
//
// Adapted from
// https://github.com/hashicorp/terraform/blob/d4ac68423c4998279f33404db46809d27a5c2362/configs/parser_values.go#L8-L23
func LoadValues(files []*hcl.File) (map[string]cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	body := hcl.MergeFiles(files)
	if body == nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Empty body in variables file",
			},
		}
	}

	vars := make(map[string]cty.Value)
	attrs, attrsDiags := body.JustAttributes()
	diags = append(diags, attrsDiags...)
	if attrs == nil {
		return vars, diags
	}

	for name, attr := range attrs {
		val, valDiags := attr.Expr.Value(nil)
		diags = append(diags, valDiags...)
		vars[name] = val
	}

	return vars, diags
}
