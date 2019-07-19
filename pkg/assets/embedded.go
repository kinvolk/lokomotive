package assets

import (
	"fmt"

	"github.com/gobuffalo/packd"
	packr "github.com/gobuffalo/packr/v2"
)

type embeddedAssets struct {
	locationsToBoxes map[string]*packr.Box
}

var _ AssetsIface = &embeddedAssets{}

func newEmbeddedAssets() *embeddedAssets {
	return &embeddedAssets{
		locationsToBoxes: map[string]*packr.Box{
			"/components/cert-manager/manifests":                  packr.New("cert-manager", "../../assets/components/cert-manager/manifests/"),
			"/components/contour/manifests-deployment":            packr.New("contour-deployment", "../../assets/components/contour/manifests-deployment/"),
			"/components/contour/manifests-daemonset":             packr.New("contour-daemonset", "../../assets/components/contour/manifests-daemonset/"),
			"/components/ingress-nginx/manifests":                 packr.New("ingress-nginx", "../../assets/components/ingress-nginx/manifests/"),
			"/components/openebs-default-storage-class/manifests": packr.New("openebs-default-storage-class", "../../assets/components/openebs-default-storage-class/manifests/"),
			"/components/prometheus-operator/manifests":           packr.New("prometheus-operator", "../../assets/components/prometheus-operator/manifests/"),

			"/lokomotive-kubernetes": packr.New("lokomotive-kubernetes", "../../assets/lokomotive-kubernetes"),
		},
	}
}

func (a *embeddedAssets) WalkFiles(location string, cb WalkFunc) error {
	box, ok := a.locationsToBoxes[location]
	if !ok {
		return fmt.Errorf("no box with assets for %q", location)
	}
	return box.Walk(func(fileName string, file packd.File) error {
		fileInfo, err := file.FileInfo()
		return cb(fileName, fileInfo, file, err)
	})
}
