package walkers

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/assets"
)

func CopyingWalker(path string, newDirPerms os.FileMode) assets.WalkFunc {
	return func(fileName string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			return errors.Wrapf(err, "error during walking at %q", fileName)
		}

		fileName = filepath.Join(path, fileName)

		if err := os.MkdirAll(filepath.Dir(fileName), newDirPerms); err != nil {
			return errors.Wrap(err, "failed to create dir")
		}

		targetFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, fileInfo.Mode())
		if err != nil {
			return errors.Wrap(err, "failed to open target file")
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, r); err != nil {
			return errors.Wrap(err, "failed to write file")
		}
		return nil
	}
}
