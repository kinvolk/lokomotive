package components

import (
	"github.com/hashicorp/hcl/v2"
)

var (
	HCLDiagConfigBodyNil = &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "*hcl.Body is nil",
		Detail:   "*hcl.Body pointer must not be nil - did you provide a lokocfg configuration file?",
	}
)
