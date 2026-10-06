package zone

import (
	"context"

	"github.com/kumahq/kuma/v3/pkg/core/tokens"
)

type Validator interface {
	Validate(ctx context.Context, token Token) (Identity, error)
}

type jwtTokenValidator struct {
	validator tokens.Validator
}

var _ Validator = &jwtTokenValidator{}

func NewValidator(validator tokens.Validator) Validator {
	return &jwtTokenValidator{
		validator: validator,
	}
}

func (j *jwtTokenValidator) Validate(ctx context.Context, token Token) (Identity, error) {
	claims := &ZoneClaims{}
	if err := j.validator.ParseWithValidation(ctx, token, claims); err != nil {
		return Identity{}, err
	}
	return Identity{
		Zone:  claims.Zone,
		Scope: claims.Scope,
	}, nil
}
