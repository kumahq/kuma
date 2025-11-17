package issuer

import (
	"github.com/golang-jwt/jwt/v4"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/tokens"
	"github.com/kumahq/kuma/v2/pkg/core/user"
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

func (c *UserClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	c.RegisteredClaims = claims
}
