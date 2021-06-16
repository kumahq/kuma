package auth

import (
	"context"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type Credential = string

type Authenticator interface {
	Authenticate(ctx context.Context, dataplane *core_mesh.DataplaneResource, credential Credential) error
	AuthenticateZoneIngress(ctx context.Context, zoneIngress *core_mesh.ZoneIngressResource, credential Credential) error
}
