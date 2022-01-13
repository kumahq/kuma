package context

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
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

type ControlPlaneContext struct {
	AdminProxyKeyPair *tls.KeyPair
	CLACache          xds.CLACache
	Secrets           secrets.Secrets
}

type MeshContext struct {
	Resource    *core_mesh.MeshResource
	Dataplanes  *core_mesh.DataplaneResourceList
	Hash        string
	EndpointMap xds.EndpointMap
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

func BuildControlPlaneContext(
	config kuma_cp.Config,
	claCache xds.CLACache,
	secrets secrets.Secrets,
) (*ControlPlaneContext, error) {
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
