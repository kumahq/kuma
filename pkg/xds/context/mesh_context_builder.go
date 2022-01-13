package context

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/dns/vips"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var logger = core.Log.WithName("xds").WithName("context")

type meshContextBuilder struct {
	ipFunc          lookup.LookupIPFunc
	zone            string
	vipsPersistence *vips.Persistence
	topLevelDomain  string
}

type MeshContextBuilder interface {
	Build(snapshot *MeshSnapshot) (MeshContext, error)
}

func NewMeshContextBuilder(ipFunc lookup.LookupIPFunc, zone string, vipsPersistence *vips.Persistence, topLevelDomain string) MeshContextBuilder {
	return &meshContextBuilder{
		ipFunc:          ipFunc,
		zone:            zone,
		vipsPersistence: vipsPersistence,
		topLevelDomain:  topLevelDomain,
	}
}

func (m *meshContextBuilder) Build(snapshot *MeshSnapshot) (MeshContext, error) {
	dataplanesList := snapshot.Resources(core_mesh.DataplaneType).(*core_mesh.DataplaneResourceList)
	zoneIngressList := snapshot.Resources(core_mesh.ZoneIngressType).(*core_mesh.ZoneIngressResourceList)
	zoneIngresses := xds_topology.ResolveZoneIngressAddresses(logger, m.ipFunc, zoneIngressList.Items)
	serviceInsights := snapshot.Resources(core_mesh.ServiceInsightType).(*core_mesh.ServiceInsightResourceList)

	dataplanes := xds_topology.ResolveAddresses(logger, m.ipFunc, dataplanesList.Items)

	dataplanesByName := map[string]*core_mesh.DataplaneResource{}
	for _, dp := range dataplanes {
		dataplanesByName[dp.Meta.GetName()] = dp
	}

	virtualOutboundView, err := m.vipsPersistence.GetByMesh(snapshot.mesh.GetMeta().GetName())
	if err != nil {
		return MeshContext{}, err
	}
	// resolve all the domains
	domains, outbounds := xds_topology.VIPOutbounds(virtualOutboundView, m.topLevelDomain)

	return MeshContext{
		Resource: snapshot.mesh,
		Dataplanes: &core_mesh.DataplaneResourceList{
			Items: dataplanes,
		},
		DataplanesByName: dataplanesByName,
		ZoneIngresses: &core_mesh.ZoneIngressResourceList{
			Items: zoneIngresses,
		},
		Hash:                snapshot.Hash,
		EndpointMap:         xds_topology.BuildEdsEndpointMap(snapshot.mesh, m.zone, dataplanes, zoneIngresses),
		VIPOutbounds:        outbounds,
		VIPDomains:          domains,
		Snapshot:            snapshot,
		ServiceTLSReadiness: resolveTLSReadiness(snapshot.mesh, serviceInsights),
	}, nil
}

func resolveTLSReadiness(mesh *core_mesh.MeshResource, serviceInsights *core_mesh.ServiceInsightResourceList) map[string]bool {
	tlsReady := map[string]bool{}

	backend := mesh.GetEnabledCertificateAuthorityBackend()
	// TLS readiness is irrelevant unless we are using PERMISSIVE TLS, so skip
	// checking ServiceInsights if we aren't.
	if backend == nil || backend.Mode != mesh_proto.CertificateAuthorityBackend_PERMISSIVE {
		return tlsReady
	}

	if len(serviceInsights.Items) == 0 {
		// Nothing about the TLS readiness has been reported yet
		logger.Info("could not determine service TLS readiness, ServiceInsight is not yet present")
		return tlsReady
	}

	for svc, insight := range serviceInsights.Items[0].Spec.GetServices() {
		tlsReady[svc] = insight.IssuedBackends[backend.Name] == insight.Dataplanes.Total
	}
	return tlsReady
}
