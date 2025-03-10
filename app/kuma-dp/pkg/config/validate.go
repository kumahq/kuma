package config

import (
	"os"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/golang-jwt/jwt/v5"
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

	rawToken, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "could not read the token in the file %s", path)
	}

	strToken := strings.TrimSpace(string(rawToken))
	if !govalidator.Matches(strToken, "^[^\\x00\\n\\r]*$") {
		return errors.New("Token shouldn't contain line breaks within the token, only at the start or end")
	}
	token, _, err := new(jwt.Parser).ParseUnverified(strToken, &jwt.MapClaims{})
	if err != nil {
		return errors.Wrap(err, "not valid JWT token. Can't parse it.")
	}

	if token.Method.Alg() == "" {
		return errors.New("not valid JWT token. No Alg.")
	}

	if token.Header == nil {
		return errors.New("not valid JWT token. No Header.")
	}

	return nil
}
