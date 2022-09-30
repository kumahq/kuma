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

type UserClaims struct {
	user.User
	jwt.RegisteredClaims
}

var _ tokens.Claims = &UserClaims{}

func (c *UserClaims) ID() string {
	return c.RegisteredClaims.ID
}

func (c *UserClaims) KeyIDFallback() (int, error) {
	return 0, errors.New("kid is required") // kid was required when we introduced User Token
}

func (c *UserClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	c.RegisteredClaims = claims
}
