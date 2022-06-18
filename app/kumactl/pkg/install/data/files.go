package data

import (
	"fmt"
	"io"
	"io/fs"
)

type FileList []File

type File struct {
	Data     []byte
	Name     string
	FullPath string
}

func (f File) String() string {
	return string(f.Data)
}

func ReadFiles(fileSys fs.FS) (FileList, error) {
	files := []File{}

	err := fs.WalkDir(fileSys, ".", func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !dir.IsDir() {
			data, err := fs.ReadFile(fileSys, path)
			if err != nil {
				return fmt.Errorf("Failed to read file: %s: %w", path, err)
			}
			file := File{
				Data:     data,
				Name:     dir.Name(),
				FullPath: path,
			}
			files = append(files, file)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func ReadFile(fileSys fs.FS, file string) (File, error) {
	f, err := fileSys.Open(file)
	if err != nil {
		return File{}, err
	}
	b, err := io.ReadAll(f)
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
