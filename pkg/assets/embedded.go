package assets

//go:generate go run assets_generate.go

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shurcooL/httpfs/vfsutil"
)

type embeddedAssets struct {
	fs http.FileSystem
}

var _ AssetsIface = &embeddedAssets{}

func newEmbeddedAssets() *embeddedAssets {
	return &embeddedAssets{
		fs: vfsgenAssets,
	}
}

func (a *embeddedAssets) WalkFiles(location string, cb WalkFunc) error {
	return vfsutil.WalkFiles(a.fs, location, func(filePath string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			return cb(filePath, fileInfo, r, err)
		}
		if fileInfo.IsDir() {
			return nil
		}
		relPath, relErr := filepath.Rel(location, filePath)
		if relErr != nil {
			return cb(relPath, fileInfo, nil, relErr)
		}
		return cb(relPath, fileInfo, r, err)
	})
}
