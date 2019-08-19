package assets

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/alertmanager/pkg/modtimevfs"
	"github.com/shurcooL/httpfs/union"
	"github.com/shurcooL/vfsgen"
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
	if value, found := os.LookupEnv("LOKOCTL_USE_FS_ASSETS"); found {
		Assets = newFsAssets(value)
	}
}

// Generate function wraps vfsgen.Generate function.
// Additionally to vfsgen.Generate, it also takes map of directories,
// where key represents path in the assets and a value represents path
// to the assets directory in local filesystem (which should be relative).
//
// This function also resets modification time for every file, so creating a new copy
// of code does not trigger changes in all asset files.
func Generate(fileName string, packageName string, variableName string, dirs map[string]string) error {
	ufs := make(map[string]http.FileSystem)
	for assetsPath, fsPath := range dirs {
		ufs[assetsPath] = http.Dir(fsPath)
	}
	u := union.New(ufs)
	fs := modtimevfs.New(u, time.Unix(1, 0))
	return vfsgen.Generate(fs, vfsgen.Options{
		Filename:     fileName,
		PackageName:  packageName,
		VariableName: variableName,
	})
}
