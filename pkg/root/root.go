package root

import (
	"path"
	"path/filepath"
	"runtime"
)

var basePath = "../"

func GetRootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b), basePath)
	return filepath.Dir(d)
}
