package issuer

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/user"
)

type Token = string

type UserTokenIssuer interface {
	Generate(identity user.User, validFor time.Duration) (Token, error)
	Validate(token Token) (user.User, error)
}

// jwtTokenIssuer generates and validates User Tokens.
// User Token is a JWT token with Zone and Serial Number.
// We place Serial Number in token so we don't have to validate the token against every single signing key.
// Instead we take Serial Number from the token, retrieve signing key and validate only against this key.
// A new token is always generated by using the latest signing key.
type jwtTokenIssuer struct {
	signingKeyManager SigningKeyManager
	revocations       TokenRevocations
}

func NewUserTokenIssuer(signingKeyAccessor SigningKeyManager, revocations TokenRevocations) UserTokenIssuer {
	return &jwtTokenIssuer{
		signingKeyManager: signingKeyAccessor,
		revocations:       revocations,
	}
}

var _ UserTokenIssuer = &jwtTokenIssuer{}

type claims struct {
	user.User
	jwt.RegisteredClaims
}

func (j *jwtTokenIssuer) Generate(identity user.User, validFor time.Duration) (Token, error) {
	signingKey, serialNumber, err := j.signingKeyManager.GetLatestSigningKey()
	if err != nil {
		return "", err
	}

	now := core.Now()
	c := claims{
		User: identity,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        core.NewUUID(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(time.Minute * -5)), // todo(jakubdyszkiewicz) parametrize via config and go through all clock skews in the project
		},
	}

	if validFor != 0 {
		c.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(now.Add(validFor))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	token.Header["kid"] = strconv.Itoa(serialNumber)
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", errors.Wrap(err, "could not sign a token")
	}
	return tokenString, nil
}

func (j *jwtTokenIssuer) Validate(rawToken Token) (user.User, error) {
	c := &claims{}
	token, err := jwt.ParseWithClaims(rawToken, c, func(token *jwt.Token) (interface{}, error) {
		serialNumberRaw := token.Header["kid"]
		if serialNumberRaw == nil {
			return nil, errors.New("kid header not found")
		}
		serialNumber, err := strconv.Atoi(serialNumberRaw.(string))
		if err != nil {
			return nil, err
		}
		signingKey, err := j.signingKeyManager.GetSigningKey(serialNumber)
		if err != nil {
			return nil, errors.Wrapf(err, "could not get signing key with serial number %d. The signing key most likely has been rotated, regenerate the token", serialNumber)
		}
		return signingKey, nil
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
