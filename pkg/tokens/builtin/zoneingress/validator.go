package zoneingress

import (
	"context"

	core_tokens "github.com/kumahq/kuma/pkg/core/tokens"
)

type Validator interface {
	Validate(ctx context.Context, token Token) (Identity, error)
}

type jwtValidator struct {
	validator core_tokens.Validator
}

var _ Validator = &jwtValidator{}

func NewValidator(validator core_tokens.Validator) Validator {
	return &jwtValidator{
		validator: validator,
	}
}

func (j *jwtValidator) Validate(ctx context.Context, token Token) (Identity, error) {
	claims := &zoneIngressClaims{}
	if err := j.validator.ParseWithValidation(ctx, token, claims); err != nil {
		return Identity{}, err
	}
	return Identity{
		Zone: claims.Zone,
	}, nil
}
