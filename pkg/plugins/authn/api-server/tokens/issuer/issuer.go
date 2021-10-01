package issuer

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/user"
)

type Token = string

type UserTokenIssuer interface {
	Generate(identity user.User, validFor time.Duration) (Token, error)
	Validate(token Token) (user.User, int, error)
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
	TokenSerialNumber int
	jwt.RegisteredClaims
}

func (j *jwtTokenIssuer) Generate(identity user.User, validFor time.Duration) (Token, error) {
	signingKey, serialNumber, err := j.signingKeyManager.GetLatestSigningKey()
	if err != nil {
		return "", err
	}

	now := core.Now()
	c := claims{
		User:              identity,
		TokenSerialNumber: serialNumber,
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
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", errors.Wrap(err, "could not sign a token")
	}
	return tokenString, nil
}

func (j *jwtTokenIssuer) Validate(rawToken Token) (user.User, int, error) {
	// first extract the token serial number without of token itself verification yet
	c := &claims{}
	_, _, err := new(jwt.Parser).ParseUnverified(rawToken, c)
	if err != nil {
		return user.User{}, 0, errors.Wrap(err, "could not parse token")
	}

	if c.TokenSerialNumber == 0 {
		return user.User{}, 0, errors.New("token has no serial number")
	}

	// get signing key of serial number in the token
	signingKey, err := j.signingKeyManager.GetSigningKey(c.TokenSerialNumber)
	if err != nil {
		return user.User{}, 0, errors.Wrapf(err, "could not get Signing Key with serial number %d. Signing Key most likely has been rotated, regenerate the token", c.TokenSerialNumber)
	}

	// verify the token
	token, err := jwt.ParseWithClaims(rawToken, c, func(*jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return user.User{}, 0, errors.Wrap(err, "could not parse token")
	}
	if !token.Valid {
		return user.User{}, 0, errors.New("token is not valid")
	}

	revoked, err := j.revocations.IsRevoked(c.ID)
	if err != nil {
		return user.User{}, 0, errors.Wrap(err, "could not check if the token is revoked")
	}
	if revoked {
		return user.User{}, 0, errors.New("token is revoked")
	}

	return c.User, c.TokenSerialNumber, nil
}
