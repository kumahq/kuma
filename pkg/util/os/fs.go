package os

import (
	"os"

	"github.com/pkg/errors"
)

func TryWriteToDir(dir string) error {
	file, err := os.CreateTemp(dir, "write-access-check")
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModeDir|0755); err != nil {
				return errors.Wrapf(err, "unable to create a directory %q", dir)
			}
			file, err = os.CreateTemp(dir, "write-access-check")
		}
		if err != nil {
			return errors.Wrapf(err, "unable to create temporary files in directory %q", dir)
		}
	}
	if err := os.Remove(file.Name()); err != nil {
		return errors.Wrapf(err, "unable to remove temporary files in directory %q", dir)
	}
	return nil
}
