package dpapi

import (
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

const PATH = "/meshtrace"

// MeshTraceDpConfig is the configuration sent from CP to DP via dynconf for MeshTrace.
type MeshTraceDpConfig struct {
	Backends []OtelBackendConfig `json:"backends"`
}

type OtelBackendConfig = core_xds.OtelPipeBackend
