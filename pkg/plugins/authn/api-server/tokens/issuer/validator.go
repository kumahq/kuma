package issuer

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/core/user"
)

type UserTokenValidator interface {
	Validate(ctx context.Context, token tokens.Token) (user.User, error)
}

func NewUserTokenValidator(validator tokens.Validator) UserTokenValidator {
	return &jwtTokenValidator{
		validator: validator,
	}
}

type jwtTokenValidator struct {
	validator tokens.Validator
}

func (j *jwtTokenValidator) Validate(ctx context.Context, rawToken tokens.Token) (user.User, error) {
	claims := &userClaims{}
	if err := j.validator.ParseWithValidation(ctx, rawToken, claims); err != nil {
		return user.User{}, err
	}
	return claims.User, nil
}
