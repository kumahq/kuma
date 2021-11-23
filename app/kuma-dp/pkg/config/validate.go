package config

import (
	"os"
	"unicode"

	"github.com/golang-jwt/jwt/v4"
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

	token, parts, err := new(jwt.Parser).ParseUnverified(string(rawToken), &jwt.MapClaims{})
	if err != nil {
		return errors.Wrap(err, "not valid JWT token. Can't parse it.")
	}

	if token.Method.Alg() == "" {
		return errors.New("not valid JWT token. No Alg.")
	}

	if token.Header == nil {
		return errors.New("not valid JWT token. No Header.")
	}
	for _, part := range parts {
		if !isASCII(part) {
			return errors.New("The file cannot have blank characters like empty lines. Example how to get rid of non-printable characters: sed -i '' '/^$/d' token.file")
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
