package walkers

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/assets"
)

func DumpingWalker(contentsMap map[string]string, allowedExts ...string) assets.WalkFunc {
	var extsMap map[string]struct{}

	if len(allowedExts) > 0 {
		extsMap = make(map[string]struct{}, len(allowedExts))
		for _, ext := range allowedExts {
			extsMap[ext] = struct{}{}
		}
	}
	return func(fileName string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			return errors.Wrapf(err, "error during walking at %q", fileName)
		}

		if extsMap != nil {
			if _, ok := extsMap[filepath.Ext(fileName)]; !ok {
				return nil
			}
		}

		contents, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrapf(err, "failed to read %q", fileName)
		}

		contentsMap[fileName] = string(contents)
		return nil
	}
}
