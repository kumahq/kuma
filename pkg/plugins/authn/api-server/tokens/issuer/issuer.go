package issuer

import (
	"context"
	"time"

	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/core/user"
)

type UserTokenIssuer interface {
	Generate(ctx context.Context, identity user.User, validFor time.Duration) (tokens.Token, error)
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

func (j *jwtTokenIssuer) Generate(ctx context.Context, identity user.User, validFor time.Duration) (tokens.Token, error) {
	c := UserClaims{
		User: identity,
	}
	return j.issuer.Generate(ctx, &c, validFor)
}

type DisabledIssuer struct{}

func (d DisabledIssuer) Generate(context.Context, user.User, time.Duration) (tokens.Token, error) {
	return "", tokens.IssuerDisabled
}

var _ UserTokenIssuer = &DisabledIssuer{}
