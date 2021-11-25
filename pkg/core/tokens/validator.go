package tokens

import (
	"strconv"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

type Validator interface {
	// ParseWithValidation parses token and fills data in provided Claims.
	ParseWithValidation(token Token, claims Claims) error
}

type jwtTokenValidator struct {
	keyAccessor SigningKeyAccessor
	revocations Revocations
}

func NewValidator(keyAccessor SigningKeyAccessor, revocations Revocations) Validator {
	return &jwtTokenValidator{
		keyAccessor: keyAccessor,
		revocations: revocations,
	}
}

var _ Validator = &jwtTokenValidator{}

func (j *jwtTokenValidator) ParseWithValidation(rawToken Token, claims Claims) error {
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		serialNumber, err := tokenSerialNumber(token, claims)
		if err != nil {
			return nil, err
		}
		if token.Method.Alg() == jwt.SigningMethodHS256.Name {
			return j.keyAccessor.GetLegacyKey(serialNumber)
		}
		return j.keyAccessor.GetPublicKey(serialNumber)
	})
	if err != nil {
		if verr, ok := err.(*jwt.ValidationError); ok { // jwt.ValidationError does not implement Unwrap() to just use errors.As
			if singingKeyErr, ok := verr.Inner.(*SigningKeyNotFound); ok {
				return singingKeyErr
			}
		}
		return errors.Wrap(err, "could not parse token")
	}
	if !token.Valid {
		return errors.New("token is not valid")
	}

	revoked, err := j.revocations.IsRevoked(claims.ID())
	if err != nil {
		return errors.Wrap(err, "could not check if the token is revoked")
	}
	if revoked {
		return errors.New("token is revoked")
	}

	return nil
}

func tokenSerialNumber(token *jwt.Token, claims Claims) (int, error) {
	serialNumberRaw := token.Header[KeyIDHeader]
	if serialNumberRaw != nil {
		serialNumberStr, ok := serialNumberRaw.(string)
		if !ok {
			return 0, errors.New("kid header is invalid. Expected string.")
		}
		return strconv.Atoi(serialNumberStr)
	}
	return claims.KeyIDFallback()
}
