package issuer

import (
	"strconv"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/user"
)

type UserTokenValidator interface {
	Validate(token Token) (user.User, error)
}

func NewUserTokenValidator(keyAccessor SigningKeyAccessor, revocations TokenRevocations) UserTokenValidator {
	return &jwtTokenValidator{
		keyAccessor: keyAccessor,
		revocations: revocations,
	}
}

type jwtTokenValidator struct {
	keyAccessor SigningKeyAccessor
	revocations TokenRevocations
}

func (j *jwtTokenValidator) Validate(rawToken Token) (user.User, error) {
	c := &claims{}
	token, err := jwt.ParseWithClaims(rawToken, c, func(token *jwt.Token) (interface{}, error) {
		serialNumberRaw := token.Header[KeyIDHeader]
		if serialNumberRaw == nil {
			return nil, errors.New("kid header not found")
		}
		serialNumberStr, ok := serialNumberRaw.(string)
		if !ok {
			return nil, errors.New("kid header is invalid. Expected string.")
		}
		serialNumber, err := strconv.Atoi(serialNumberStr)
		if err != nil {
			return nil, err
		}
		key, err := j.keyAccessor.GetSigningPublicKey(serialNumber)
		if err == SigningKeyNotFound {
			return nil, errors.Errorf("signing key with serial number %d not found. The signing key most likely has been rotated, regenerate the token", serialNumber)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "could not get signing key with serial number %d", serialNumber)
		}
		return key, nil
	})
	if err != nil {
		return user.User{}, errors.Wrap(err, "could not parse token")
	}
	if !token.Valid {
		return user.User{}, errors.New("token is not valid")
	}

	revoked, err := j.revocations.IsRevoked(c.ID)
	if err != nil {
		return user.User{}, errors.Wrap(err, "could not check if the token is revoked")
	}
	if revoked {
		return user.User{}, errors.New("token is revoked")
	}

	return c.User, nil
}
