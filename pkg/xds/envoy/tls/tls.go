package tls

const CpValidationCtx = "cp_validation_ctx"

// KumaALPNProtocols are set for UpstreamTlsContext to show that mTLS is created by mesh.
// On the inbound side we have to distinguish Kuma mTLS and application TLS to properly
// support PERMISSIVE mode
var KumaALPNProtocols = []string{"kuma"}
