package issuer

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_tokens "github.com/kumahq/kuma/pkg/core/tokens"
)

type Validator interface {
	Validate(ctx context.Context, token core_tokens.Token, meshName string) (DataplaneIdentity, error)
}

type jwtValidator struct {
	validators func(string) core_tokens.Validator
}

var _ Validator = &jwtValidator{}

func NewValidator(validators func(string) core_tokens.Validator) Validator {
	return &jwtValidator{
		validators: validators,
	}
}

func (j *jwtValidator) Validate(ctx context.Context, token core_tokens.Token, meshName string) (DataplaneIdentity, error) {
	claims := &DataplaneClaims{}
	if err := j.validators(meshName).ParseWithValidation(ctx, token, claims); err != nil {
		return DataplaneIdentity{}, err
	}
	return DataplaneIdentity{
		Name: claims.Name,
		Mesh: claims.Mesh,
		Tags: mesh_proto.MultiValueTagSetFrom(claims.Tags),
		Type: mesh_proto.ProxyType(claims.Type),
	}, nil
}
