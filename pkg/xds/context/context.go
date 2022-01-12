package context

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
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
	Resource      *core_mesh.MeshResource          // todo remove resource? You can access it via MeshContext.Snapshot.Mesh()
	Dataplanes    *core_mesh.DataplaneResourceList // todo remove resource? You can access it via MeshContext.Snapshot.Resources(DataplaneType) but this list has IP resolved
	ZoneIngresses *core_mesh.ZoneIngressResourceList
	Snapshot      MeshSnapshot
	Hash          string // todo technically we can get rid of this if Snapshot does not count Hash every time
	EndpointMap   xds.EndpointMap
	VIPDomains    []xds.VIPDomains
	VIPOutbounds  []*mesh_proto.Dataplane_Networking_Outbound
}

type MeshSnapshot interface {
	Hash() string
	Mesh() *core_mesh.MeshResource
	Resources(core_model.ResourceType) core_model.ResourceList
	Resource(core_model.ResourceType, core_model.ResourceKey) (core_model.Resource, bool)
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

func (mc *MeshContext) Dataplane(name string) (*core_mesh.DataplaneResource, bool) {
	for _, dp := range mc.Dataplanes.Items {
		if dp.Meta.GetName() == name {
			return dp, true
		}
	}
	return nil, false
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
	key := core_model.ResourceKey{ // we cannot use insights.ServiceInsightKey because there is an import cycle
		Name: fmt.Sprintf("all-services-%s", mc.Resource.Meta.GetName()),
		Mesh: mc.Resource.Meta.GetName(),
	}
	res, found := mc.Snapshot.Resource(core_mesh.ServiceInsightType, key)
	if !found {
		return nil, false
	}
	return res.(*core_mesh.ServiceInsightResource), true
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
