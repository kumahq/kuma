package util

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

func ClientIDFromIncomingCtx(ctx context.Context) (string, error) {
	return MetadataFromIncomingCtx(ctx, "client-id")
}

func MetadataFromIncomingCtx(ctx context.Context, key string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata is not provided")
	}
	if len(md[key]) == 0 {
		return "", errors.New("'client-id' is not present in metadata")
	}
	return md[key][0], nil
}
