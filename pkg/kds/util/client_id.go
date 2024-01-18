package util

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const clientIDKey = "client-id"

// ClientIDFromIncomingCtx returns the ID of the peer. Global has the ID
// "global" while zones have the zone name. This is also known as the peer ID.
func ClientIDFromIncomingCtx(ctx context.Context) (string, error) {
	return MetadataFromIncomingCtx(ctx, clientIDKey)
}

func MetadataFromIncomingCtx(ctx context.Context, key string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata is not provided")
	}
	if len(md[key]) == 0 {
		return "", errors.New(fmt.Sprintf("%q is not present in metadata", key))
	}
	return md[key][0], nil
}
