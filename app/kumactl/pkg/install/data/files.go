package data

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/shurcooL/httpfs/vfsutil"
)

type FileList []File

type File struct {
	Data []byte
	Name string
}

func (f File) String() string {
	return string(f.Data)
}

func ReadFiles(fs http.FileSystem) (FileList, error) {
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
			file := File{
				Data: data,
				Name: fi.Name(),
			}
			files = append(files, file)
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
		return File{}, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return File{}, err
	}
	return File{
		Data: b,
		Name: file,
	}, nil
}

func (l FileList) Filter(predicate func(File) bool) FileList {
	var list FileList
	for _, file := range l {
		if predicate(file) {
			list = append(list, file)
		}
	}
	return list
}
