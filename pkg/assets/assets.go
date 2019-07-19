package assets

import (
	"io"
	"os"
)

type WalkFunc func(fileName string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error

type AssetsIface interface {
	// WalkFiles calls cb for every regular file within path.
	//
	// Usually, fileName passed to the cb will be relative to
	// path. But in case of error, it is possible that it will
	// not.be relative. Also, in case of error, fileInfo or r may
	// be nil.
	WalkFiles(path string, cb WalkFunc) error
}

var Assets AssetsIface

func init() {
	Assets = newEmbeddedAssets()
}
