package util

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

func ClientIDFromIncomingCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata is not provided")
	}
	if len(md["client-id"]) == 0 {
		return "", errors.New("'client-id' is not present in metadata")
	}
	return md["client-id"][0], nil
}
