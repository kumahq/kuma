package issuer

import (
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/core/user"
)

type UserTokenValidator interface {
	Validate(token tokens.Token) (user.User, error)
}

func NewUserTokenValidator(validator tokens.Validator) UserTokenValidator {
	return &jwtTokenValidator{
		validator: validator,
	}
}

type jwtTokenValidator struct {
	validator tokens.Validator
}

func (j *jwtTokenValidator) Validate(rawToken tokens.Token) (user.User, error) {
	claims := &userClaims{}
	if err := j.validator.ParseWithValidation(rawToken, claims); err != nil {
		return user.User{}, err
	}
	return claims.User, nil
}
