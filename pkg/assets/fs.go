package assets

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kardianos/osext"
)

type fsAssets struct {
	assetsDir string
}

var _ AssetsIface = &fsAssets{}

func newFsAssets(dir string) *fsAssets {
	if dir == "" {
		execDir, err := osext.ExecutableFolder()
		if err != nil {
			panic("Unable to get a directory of an executable for assets")
		}
		dir = filepath.Join(execDir, "assets")
	}
	return &fsAssets{
		assetsDir: dir,
	}
}

func (a *fsAssets) WalkFiles(location string, cb WalkFunc) error {
	relativeLocation := strings.TrimLeft(location, string(os.PathSeparator))
	assetsLocation := filepath.Join(a.assetsDir, relativeLocation)
	return filepath.Walk(assetsLocation, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return cb(path, info, nil, err)
		}
		if info.IsDir() {
			return nil
		}
		relPath, relErr := filepath.Rel(assetsLocation, path)
		if relErr != nil {
			return cb(relPath, info, nil, relErr)
		}
		file, err := os.Open(path)
		return cb(relPath, info, file, err)
	})
}
