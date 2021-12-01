package tokens

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	util_rsa "github.com/kumahq/kuma/pkg/util/rsa"
)

func NewSigningKey() ([]byte, error) {
	key, err := util_rsa.GenerateKey(util_rsa.DefaultKeySize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate RSA key")
	}
	return util_rsa.ToPEMBytes(key)
}

func SigningKeyResourceKey(signingKeyPrefix string, serialNumber int, mesh string) model.ResourceKey {
	name := fmt.Sprintf("%s-%d", signingKeyPrefix, serialNumber)
	if serialNumber == 0 { // backwards compatibility with 1.3.x signing keys
		name = signingKeyPrefix
	}
	return model.ResourceKey{
		Name: name,
		Mesh: mesh,
	}
}

type SigningKeyNotFound struct {
	SerialNumber int
	Prefix       string
	Mesh         string
}

func (s *SigningKeyNotFound) Error() string {
	key := SigningKeyResourceKey(s.Prefix, s.SerialNumber, s.Mesh)
	if s.Mesh == "" {
		return fmt.Sprintf("there is no signing key with serial number %d. GlobalSecret of name %q is not found. If signing key was rotated, regenerate the token", s.SerialNumber, key.Name)
	} else {
		return fmt.Sprintf("there is no signing key with serial number %d. Secret of name %q in mesh %q is not found. If signing key was rotated, regenerate the token", s.SerialNumber, key.Name, key.Mesh)
	}
}

func IsSigningKeyNotFound(err error) bool {
	target := &SigningKeyNotFound{}
	return errors.As(err, &target)
}

func keyBytesToRsaKey(keyBytes []byte) (*rsa.PrivateKey, error) {
	if util_rsa.IsPEMBytes(keyBytes) {
		key, err := util_rsa.FromPEMBytes(keyBytes)
		if err != nil {
			return nil, err
		}
		return key, nil
	}

	// support non-PEM RSA key for legacy reasons
	key, err := x509.ParsePKCS1PrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func signingKeySerialNumber(secretName string, signingKeyPrefix string) (int, error) {
	serialNumberStr := strings.ReplaceAll(secretName, signingKeyPrefix+"-", "")
	serialNumber, err := strconv.Atoi(serialNumberStr)
	if err != nil {
		return 0, err
	}
	return serialNumber, nil
}
