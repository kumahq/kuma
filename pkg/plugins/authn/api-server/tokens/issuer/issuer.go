package issuer

import (
	"time"

	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/core/user"
)

type UserTokenIssuer interface {
	Generate(identity user.User, validFor time.Duration) (tokens.Token, error)
}

type jwtTokenIssuer struct {
	issuer tokens.Issuer
}

func NewUserTokenIssuer(issuer tokens.Issuer) UserTokenIssuer {
	return &jwtTokenIssuer{
		issuer: issuer,
	}
}

var _ UserTokenIssuer = &jwtTokenIssuer{}

func (j *jwtTokenIssuer) Generate(identity user.User, validFor time.Duration) (tokens.Token, error) {
	c := userClaims{
		User: identity,
	}
	return j.issuer.Generate(&c, validFor)
}
