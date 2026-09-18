package auth

import (
	"context"

	"google.golang.org/grpc/credentials"
)

type tokenCredentials struct {
	loadToken                func() (string, error)
	requireTransportSecurity bool
}

var _ credentials.PerRPCCredentials = &tokenCredentials{}

// NewTokenCredentials attaches the token to every KDS RPC. loadToken is called
// on every RPC, so a token backed by a file is picked up on the next stream.
// With requireTransportSecurity the connection is refused unless it is TLS.
func NewTokenCredentials(loadToken func() (string, error), requireTransportSecurity bool) credentials.PerRPCCredentials {
	return &tokenCredentials{
		loadToken:                loadToken,
		requireTransportSecurity: requireTransportSecurity,
	}
}

func (t *tokenCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	token, err := t.loadToken()
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, nil
	}
	return map[string]string{TokenHeader: token}, nil
}

func (t *tokenCredentials) RequireTransportSecurity() bool {
	return t.requireTransportSecurity
}
