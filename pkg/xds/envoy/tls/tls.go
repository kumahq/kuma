package tls

import (
	"fmt"
)

const CpValidationCtx = "cp_validation_ctx"

// KumaALPNProtocols are set for UpstreamTlsContext to show that mTLS is created by mesh.
// On the inbound side we have to distinguish Kuma mTLS and application TLS to properly
// support PERMISSIVE mode
var KumaALPNProtocols = []string{"kuma"}

func ServiceSpiffeID(mesh string, service string) string {
	return fmt.Sprintf("spiffe://%s/%s", mesh, service)
}
