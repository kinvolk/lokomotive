package util

import (
	"fmt"
	"testing"
)

func TestRenderChartBadValues(t *testing.T) {
	c := "cert-manager"
	values := "malformed\t"

	helmChart, err := LoadChartFromAssets(fmt.Sprintf("/components/%s/manifests", c))
	if err != nil {
		t.Fatalf("Loading chart from assets should succeed, got: %v", err)
	}

	if _, err := RenderChart(helmChart, c, c, values); err == nil {
		t.Fatalf("Rendering chart with malformed values should fail")
	}
}
