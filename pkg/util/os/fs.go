package os

import (
	"fmt"
	"os"
)

func TryWriteToDir(dir string) error {
	file, err := os.CreateTemp(dir, "write-access-check")
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModeDir|0o755); err != nil {
				return fmt.Errorf("unable to create a directory %q: %w", dir, err)
			}
			file, err = os.CreateTemp(dir, "write-access-check")
		}
		if err != nil {
			return fmt.Errorf("unable to create temporary files in directory %q: %w", dir, err)
		}
	}
	if err := os.Remove(file.Name()); err != nil {
		return fmt.Errorf("unable to remove temporary files in directory %q: %w", dir, err)
	}
	return nil
}
