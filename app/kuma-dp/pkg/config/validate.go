package config

import (
	"io/ioutil"
	"unicode"

	"github.com/kumahq/kuma/pkg/core"

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

	token, parts, err := new(jwt.Parser).ParseUnverified(string(rawToken), &jwt.MapClaims{})
	if err != nil {
		return errors.Wrapf(err, "token in the file %s is not valid JWT token. Double check for blank characters in the file", path)
	}

	core.Log.Info("dump ", "token", token, "parts ", parts)
	if token.Method.Alg() == "" {
		return errors.New("not valid JWT token. No Alg.")
	}

	if token.Header == nil {
		return errors.New("not valid JWT token. No Header.")
	}
	if len(parts) != 3 {
		return errors.New("not valid JWT token. Must have 3 parts..")
	}
	for _, part := range parts {
		if !isASCII(part) {
			return errors.New("token in the file %s is not valid JWT token. Found a non-printable characters in the token.")
		}
	}

	return nil
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if !unicode.IsPrint(rune(s[i])) {
			return false
		}
	}
	return true
}
