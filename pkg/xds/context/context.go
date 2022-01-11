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
	Resource         *core_mesh.MeshResource
	Dataplanes       *core_mesh.DataplaneResourceList
	DataplanesByName map[string]*core_mesh.DataplaneResource
	ZoneIngresses    *core_mesh.ZoneIngressResourceList
	Snapshot         *MeshSnapshot
	Hash             string
	EndpointMap      xds.EndpointMap
	VIPDomains       []xds.VIPDomains
	VIPOutbounds     []*mesh_proto.Dataplane_Networking_Outbound
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

func (mc *MeshContext) ExternalServices() *core_mesh.ExternalServiceResourceList {
	return mc.Snapshot.Resources(core_mesh.ExternalServiceType).(*core_mesh.ExternalServiceResourceList)
}

func (mc *MeshContext) HealthChecks() *core_mesh.HealthCheckResourceList {
	return mc.Snapshot.Resources(core_mesh.HealthCheckType).(*core_mesh.HealthCheckResourceList)
}

func (mc *MeshContext) TrafficTraces() *core_mesh.TrafficTraceResourceList {
	return mc.Snapshot.Resources(core_mesh.TrafficTraceType).(*core_mesh.TrafficTraceResourceList)
}

func (mc *MeshContext) TrafficRoutes() *core_mesh.TrafficRouteResourceList {
	return mc.Snapshot.Resources(core_mesh.TrafficRouteType).(*core_mesh.TrafficRouteResourceList)
}

func (mc *MeshContext) Retries() *core_mesh.RetryResourceList {
	return mc.Snapshot.Resources(core_mesh.RetryType).(*core_mesh.RetryResourceList)
}

func (mc *MeshContext) TrafficPermissions() *core_mesh.TrafficPermissionResourceList {
	return mc.Snapshot.Resources(core_mesh.TrafficPermissionType).(*core_mesh.TrafficPermissionResourceList)
}

func (mc *MeshContext) TrafficLogs() *core_mesh.TrafficLogResourceList {
	return mc.Snapshot.Resources(core_mesh.TrafficLogType).(*core_mesh.TrafficLogResourceList)
}

func (mc *MeshContext) FaultInjections() *core_mesh.FaultInjectionResourceList {
	return mc.Snapshot.Resources(core_mesh.FaultInjectionType).(*core_mesh.FaultInjectionResourceList)
}

func (mc *MeshContext) Timeouts() *core_mesh.TimeoutResourceList {
	return mc.Snapshot.Resources(core_mesh.TimeoutType).(*core_mesh.TimeoutResourceList)
}

func (mc *MeshContext) RateLimits() *core_mesh.RateLimitResourceList {
	return mc.Snapshot.Resources(core_mesh.RateLimitType).(*core_mesh.RateLimitResourceList)
}

func (mc *MeshContext) CircuitBreakers() *core_mesh.CircuitBreakerResourceList {
	return mc.Snapshot.Resources(core_mesh.CircuitBreakerType).(*core_mesh.CircuitBreakerResourceList)
}

func (mc *MeshContext) ServiceInsight() (*core_mesh.ServiceInsightResource, bool) {
	resources := mc.Snapshot.Resources(core_mesh.ServiceInsightType).(*core_mesh.ServiceInsightResourceList)
	if len(resources.Items) > 0 {
		return resources.Items[0], true // there is only one service insight for mesh
	}
	return nil, false
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
