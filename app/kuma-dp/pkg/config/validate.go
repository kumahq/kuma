package config

import (
	"github.com/pkg/errors"

	util_files "github.com/Kong/kuma/pkg/util/files"
)

func ValidateTokenPath(path string) error {
	if path == "" {
		return nil
	}
	empty, err := util_files.FileEmpty(path)
	if err != nil {
		return errors.Wrap(err, "could not read file")
	}
	if empty {
		return errors.Errorf("token under file %s is empty", path)
	}
	return nil
}
