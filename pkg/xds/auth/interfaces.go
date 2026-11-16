package auth

import (
	"context"

	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

type Credential = string

type Authenticator interface {
	// Authenticate verifies the credential for the resource and returns which of
	// the resource's labels the credential vouches for.
	Authenticate(ctx context.Context, resource model.Resource, credential Credential) (Binding, error)
}
