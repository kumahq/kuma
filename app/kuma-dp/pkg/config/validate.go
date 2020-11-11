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
		return errors.Wrapf(err, "could not read file %s", path)
	}
	if empty {
		return errors.Errorf("token under file %s is empty", path)
	}

	rawToken, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "could not read the token in the file %s", path)
	}

	_, _, err = new(jwt.Parser).ParseUnverified(string(rawToken), &jwt.MapClaims{})
	if err != nil {
		return errors.Wrapf(err, "token in the file %s is not valid JWT token. Double check blank characters in the file", path)
	}

	return nil
}
