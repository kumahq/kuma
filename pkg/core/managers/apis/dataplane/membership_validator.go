package dataplane

import (
	"context"

	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

type membershipValidator struct{}

var _ Validator = &membershipValidator{}

func NewMembershipValidator() Validator {
	return &membershipValidator{}
}

func (m *membershipValidator) ValidateCreate(_ context.Context, _ model.ResourceKey, _ *core_mesh.DataplaneResource, _ *core_mesh.MeshResource) error {
	return nil
}

func (m *membershipValidator) ValidateUpdate(_ context.Context, _ *core_mesh.DataplaneResource, _ *core_mesh.MeshResource) error {
	return nil
}
