package auth

import (
	"context"

	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

type Credential = string

type Authenticator interface {
	Authenticate(ctx context.Context, resource model.Resource, credential Credential) error
}
