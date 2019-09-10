package data

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/shurcooL/httpfs/vfsutil"
)

type File []byte

func (f File) String() string {
	return string(f)
}

func ReadFiles(fs http.FileSystem) ([]File, error) {
	files := []File{}

	walkFn := func(path string, fi os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			data, err := ioutil.ReadAll(r)
			if err != nil {
				return errors.Wrapf(err, "Failed to read file: %s", path)
			}
			files = append(files, data)
		}
		return nil
	}

	err := vfsutil.WalkFiles(fs, "/", walkFn)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func ReadFile(fs http.FileSystem, file string) (File, error) {
	f, err := fs.Open(file)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return b, nil
}
