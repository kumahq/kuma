package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin/zone"
)

var log = core.Log.WithName("kds").WithName("auth")

type zoneTokenAuthenticator struct {
	validator zone.Validator
}

var _ Authenticator = &zoneTokenAuthenticator{}

// NewZoneTokenAuthenticator accepts Zone CPs presenting a Zone Token issued for their zone with the "cp" scope.
func NewZoneTokenAuthenticator(validator zone.Validator) Authenticator {
	return &zoneTokenAuthenticator{validator: validator}
}

func (z *zoneTokenAuthenticator) Authenticate(ctx context.Context, md metadata.MD, zoneName string) error {
	token, err := extractToken(md, zoneName)
	if err != nil {
		return err
	}
	identity, err := z.validator.Validate(ctx, token)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "invalid zone token: %s", err)
	}
	if identity.Zone != zoneName {
		return status.Errorf(codes.PermissionDenied, "token is signed for %q zone, but connected CP advertised as %q", identity.Zone, zoneName)
	}
	if !zone.InScope(identity.Scope, zone.CPScope) {
		return status.Errorf(codes.PermissionDenied, "token cannot be used to authenticate zone CP (%s is out of token's scope: %+v)", zone.CPScope, identity.Scope)
	}
	log.V(1).Info("zone CP authenticated", "zone", zoneName)
	return nil
}

func extractToken(md metadata.MD, zoneName string) (string, error) {
	values := md.Get(TokenHeader)
	switch len(values) {
	case 0:
		return "", status.Errorf(codes.Unauthenticated,
			"Zone CP did not provide a zone token. 1) Generate the token using `kumactl generate zone-token --zone=%s --scope=%s` on Global CP. 2) Set the token in KUMA_MULTIZONE_ZONE_KDS_AUTH_TOKEN_INLINE or as a file in KUMA_MULTIZONE_ZONE_KDS_AUTH_TOKEN_PATH on Zone CP.",
			zoneName, zone.CPScope,
		)
	case 1:
		token := strings.TrimPrefix(values[0], "Bearer ")
		return strings.TrimPrefix(token, "bearer "), nil
	default:
		return "", status.Errorf(codes.Unauthenticated, "%s header has %d values. Expected 1 value", TokenHeader, len(values))
	}
}
