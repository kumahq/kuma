package mesh

import (
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/dns/vips"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var logger = core.Log.WithName("xds").WithName("mesh")

type meshContextBuilder struct {
	ipFunc          lookup.LookupIPFunc
	zone            string
	vipsPersistence *vips.Persistence
	topLevelDomain  string
}

type MeshContextBuilder interface { // todo move to xds/context
	Build(snapshot xds_context.MeshSnapshot) (xds_context.MeshContext, error)
}

func NewMeshContextBuilder(ipFunc lookup.LookupIPFunc, zone string, vipsPersistence *vips.Persistence, topLevelDomain string) MeshContextBuilder {
	return &meshContextBuilder{
		ipFunc:          ipFunc,
		zone:            zone,
		vipsPersistence: vipsPersistence,
		topLevelDomain:  topLevelDomain,
	}
}

func (m *meshContextBuilder) Build(snapshot xds_context.MeshSnapshot) (xds_context.MeshContext, error) {
	dataplanesList := snapshot.Resources(core_mesh.DataplaneType).(*core_mesh.DataplaneResourceList)
	dataplanes := xds_topology.ResolveAddresses(logger, m.ipFunc, dataplanesList.Items)

	zoneIngressList := snapshot.Resources(core_mesh.ZoneIngressType).(*core_mesh.ZoneIngressResourceList)
	zoneIngresses := xds_topology.ResolveZoneIngressAddresses(logger, m.ipFunc, zoneIngressList.Items)

	virtualOutboundView, err := m.vipsPersistence.GetByMesh(snapshot.Mesh().GetMeta().GetName())
	if err != nil {
		return xds_context.MeshContext{}, err
	}
	// resolve all the domains
	domains, outbounds := xds_topology.VIPOutbounds(virtualOutboundView, m.topLevelDomain)

	return xds_context.MeshContext{
		Resource: snapshot.Mesh(),
		Dataplanes: &core_mesh.DataplaneResourceList{
			Items: dataplanes,
		},
		ZoneIngresses: &core_mesh.ZoneIngressResourceList{
			Items: zoneIngresses,
		},
		Hash:         snapshot.Hash(),
		EndpointMap:  xds_topology.BuildEdsEndpointMap(snapshot.Mesh(), m.zone, dataplanes, zoneIngresses),
		VIPOutbounds: outbounds,
		VIPDomains:   domains,
		Snapshot:     snapshot,
	}, nil
}
