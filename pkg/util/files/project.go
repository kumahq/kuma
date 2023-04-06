package files

import (
	"go/build"
	"os"
	"path"
	"strings"
)

func GetProjectRoot(file string) string {
	dir := file
	for path.Base(dir) != "pkg" && path.Base(dir) != "app" {
		dir = path.Dir(dir)
	}
	return path.Dir(dir)
}

func GetProjectRootParent(file string) string {
	return path.Dir(GetProjectRoot(file))
}

func RelativeToPkgMod(file string) string {
	root := path.Dir(path.Dir(path.Dir(GetProjectRoot(file))))
	return strings.TrimPrefix(file, root)
}

func RelativeToProjectRoot(path string) string {
	root := GetProjectRoot(path)
	return strings.TrimPrefix(path, root)
}

func RelativeToProjectRootParent(path string) string {
	root := GetProjectRootParent(path)
	return strings.TrimPrefix(path, root)
}

func GetGopath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	return gopath
}
