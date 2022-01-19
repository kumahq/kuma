package context

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/tls"
	"github.com/kumahq/kuma/pkg/xds/secrets"
)

type Context struct {
	ControlPlane *ControlPlaneContext
	Mesh         MeshContext
}

type ConnectionInfo struct {
	// Authority defines the URL that was used by the data plane to connect to the control plane
	Authority string
}

// ControlPlaneContext contains shared global data and components that are required for generating XDS
// This data is the same regardless of a data plane proxy and mesh we are generating the data for.
type ControlPlaneContext struct {
	AdminProxyKeyPair *tls.KeyPair
	CLACache          xds.CLACache
	Secrets           secrets.Secrets
}

// MeshContext contains shared data within one mesh that is required for generating XDS config.
// This data is the same for all data plane proxies within one mesh.
// If there is an information that can be precomputed and shared between all data plane proxies
// it should be put here. This way we can save CPU cycles of computing the same information.
type MeshContext struct {
	Hash                string
	Resource            *core_mesh.MeshResource
	Resources           Resources
	DataplanesByName    map[string]*core_mesh.DataplaneResource
	EndpointMap         xds.EndpointMap
	VIPDomains          []xds.VIPDomains
	VIPOutbounds        []*mesh_proto.Dataplane_Networking_Outbound
	ServiceTLSReadiness map[string]bool
}

func (mc *MeshContext) GetTracingBackend(tt *core_mesh.TrafficTraceResource) *mesh_proto.TracingBackend {
	if tt == nil {
		return nil
	}
	if tb := mc.Resource.GetTracingBackend(tt.Spec.GetConf().GetBackend()); tb == nil {
		core.Log.Info("Tracing backend is not found. Ignoring.",
			"backendName", tt.Spec.GetConf().GetBackend(),
			"trafficTraceName", tt.GetMeta().GetName(),
			"trafficTraceMesh", tt.GetMeta().GetMesh())
		return nil
	} else {
		return tb
	}
}

func (mc *MeshContext) GetLoggingBackend(tl *core_mesh.TrafficLogResource) *mesh_proto.LoggingBackend {
	if tl == nil {
		return nil
	}
	if lb := mc.Resource.GetLoggingBackend(tl.Spec.GetConf().GetBackend()); lb == nil {
		core.Log.Info("Logging backend is not found. Ignoring.",
			"backendName", tl.Spec.GetConf().GetBackend(),
			"trafficLogName", tl.GetMeta().GetName(),
			"trafficLogMesh", tl.GetMeta().GetMesh())
		return nil
	} else {
		return lb
	}
}

func BuildControlPlaneContext(claCache xds.CLACache, secrets secrets.Secrets) (*ControlPlaneContext, error) {
	adminKeyPair, err := tls.NewSelfSignedCert("admin", tls.ServerCertType, tls.DefaultKeyType, "localhost")
	if err != nil {
		return nil, err
	}

	return &ControlPlaneContext{
		AdminProxyKeyPair: &adminKeyPair,
		CLACache:          claCache,
		Secrets:           secrets,
	}, nil
}
