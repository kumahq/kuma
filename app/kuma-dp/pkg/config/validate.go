package config

import (
	"io/ioutil"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	util_files "github.com/kumahq/kuma/pkg/util/files"
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

	rawToken, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	_, _, err = new(jwt.Parser).ParseUnverified(string(rawToken), &jwt.MapClaims{})
	if err != nil {
		return err
	}

	return nil
}
