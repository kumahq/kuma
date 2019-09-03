package auth

import (
	"context"

	"github.com/pkg/errors"

	"google.golang.org/grpc/metadata"
)

const (
	authorization = "authorization"
)

func ExtractCredential(ctx context.Context) (Credential, error) {
	metadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.Errorf("SDS request has no metadata")
	}
	if values, ok := metadata[authorization]; ok {
		if len(values) != 1 {
			return "", errors.Errorf("SDS request must have exactly 1 %q header, got %d", authorization, len(values))
		}
		return Credential(values[0]), nil
	}
	return "", nil
}
