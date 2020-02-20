package util

import (
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
