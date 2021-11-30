package tokens

import (
	"context"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

type Validator interface {
	// ParseWithValidation parses token and fills data in provided Claims.
	ParseWithValidation(ctx context.Context, token Token, claims Claims) error
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

func (j *jwtTokenValidator) ParseWithValidation(ctx context.Context, rawToken Token, claims Claims) error {
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		serialNumber, err := tokenSerialNumber(token, claims)
		if err != nil {
			return nil, err
		}
		switch token.Method.Alg() {
		case jwt.SigningMethodHS256.Name:
			return j.keyAccessor.GetLegacyKey(ctx, serialNumber)
		case jwt.SigningMethodRS256.Name:
			return j.keyAccessor.GetPublicKey(ctx, serialNumber)
		default:
			return nil, errors.Errorf("unknown token alg. Allowed algs are %s and %s", jwt.SigningMethodHS256.Name, jwt.SigningMethodRS256.Name)
		}
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

	revoked, err := j.revocations.IsRevoked(ctx, claims.ID())
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
