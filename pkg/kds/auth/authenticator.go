package auth

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// TokenHeader is the gRPC metadata key carrying the credentials of a Zone CP.
const TokenHeader = "authorization"

// Authenticator decides whether a Zone CP advertising itself as zone can use KDS.
// A gRPC status error is returned to the Zone CP as is, any other error becomes Unauthenticated.
type Authenticator interface {
	Authenticate(ctx context.Context, md metadata.MD, zone string) error
}
