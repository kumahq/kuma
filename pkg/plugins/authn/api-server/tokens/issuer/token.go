package issuer

import (
	"errors"

	"github.com/golang-jwt/jwt/v4"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/core/user"
)

const UserTokenSigningKeyPrefix = "user-token-signing-key"

var UserTokenRevocationsGlobalSecretKey = core_model.ResourceKey{
	Name: "user-token-revocations",
	Mesh: core_model.NoMesh,
}

type userClaims struct {
	user.User
	jwt.RegisteredClaims
}

var _ tokens.Claims = &userClaims{}

func (c *userClaims) ID() string {
	return c.RegisteredClaims.ID
}

func (c *userClaims) KeyIDFallback() (int, error) {
	return 0, errors.New("kid is required") // kid was required when we introduced User Token
}

func (c *userClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	c.RegisteredClaims = claims
}
