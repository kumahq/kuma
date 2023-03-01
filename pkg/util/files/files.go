package files

import (
	"io/fs"
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FileEmpty(path string) (bool, error) {
	file, err := os.Stat(path)
	if err != nil {
		return true, err
	}
	return file.Size() == 0, nil
}

// IsDirWriteable checks if dir is writable by writing and removing a file
// to dir. It returns nil if dir is writable.
func IsDirWriteable(dir string) error {
	f := filepath.Join(dir, ".touch")
	perm := 0o600
	if err := os.WriteFile(f, []byte(""), fs.FileMode(perm)); err != nil {
		return err
	}
	return os.Remove(f)
}
