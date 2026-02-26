package dpapi

import (
	policies_xds "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/xds"
)

const PATH = "/meshaccesslog"

// MeshAccessLogDpConfig is the configuration sent from CP to DP via dynconf for MeshAccessLog.
type MeshAccessLogDpConfig struct {
	Backends []OtelBackendConfig `json:"backends"`
}

type OtelBackendConfig = policies_xds.OtelBackendConfig
