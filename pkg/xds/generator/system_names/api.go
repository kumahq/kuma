package system_names

import "github.com/kumahq/kuma/pkg/core/system_names"

var (
	SystemResourceNameCABundle   = system_names.MustBeSystemName("trust_bundle")
	SystemResourceNameEnvoyAdmin = system_names.MustBeSystemName("envoy_admin")
	SystemResourceNameReadiness  = system_names.MustBeSystemName("probe_readiness")
)
