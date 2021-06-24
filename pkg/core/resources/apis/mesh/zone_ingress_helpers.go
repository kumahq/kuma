package mesh

import (
	"net"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func (r *ZoneIngressResource) UsesInboundInterface(address net.IP, port uint32) bool {
	if r == nil {
		return false
	}
	if port == r.Spec.GetNetworking().GetPort() && overlap(address, net.ParseIP(r.Spec.GetNetworking().GetAddress())) {
		return true
	}
	if port == r.Spec.GetNetworking().GetAdvertisedPort() && overlap(address, net.ParseIP(r.Spec.GetNetworking().GetAdvertisedAddress())) {
		return true
	}
	return false
}

func (r *ZoneIngressResource) IsRemoteIngress(localZone string) bool {
	if r.Spec.GetZone() == "" || r.Spec.GetZone() == localZone {
		return false
	}
	return true
}

func (r *ZoneIngressResource) HasPublicAddress() bool {
	if r == nil {
		return false
	}
	return r.Spec.GetNetworking().GetAdvertisedAddress() != "" && r.Spec.GetNetworking().GetAdvertisedPort() != 0
}

func NewZoneIngressResourceFromDataplane(dataplane *DataplaneResource) (*ZoneIngressResource, error) {
	spec, err := convert(dataplane.Spec)
	if err != nil {
		return nil, err
	}
	return &ZoneIngressResource{
		Meta: dataplane.Meta,
		Spec: spec,
	}, nil
}

func convert(dataplane *mesh_proto.Dataplane) (*mesh_proto.ZoneIngress, error) {
	if !dataplane.IsIngress() {
		return nil, errors.New("provided dataplane is not an ingress")
	}
	if len(dataplane.GetNetworking().Inbound) == 0 {
		return nil, errors.New("provided dataplane is not an ingress")
	}
	var availableServices []*mesh_proto.ZoneIngress_AvailableService
	for _, as := range dataplane.GetNetworking().GetIngress().GetAvailableServices() {
		availableServices = append(availableServices, &mesh_proto.ZoneIngress_AvailableService{
			Tags:      as.GetTags(),
			Instances: as.GetInstances(),
			Mesh:      as.GetMesh(),
		})
	}
	return &mesh_proto.ZoneIngress{
		Networking: &mesh_proto.ZoneIngress_Networking{
			Address: dataplane.GetNetworking().GetAddress(),
			Port:    dataplane.GetNetworking().Inbound[0].GetPort(),
		},
		AvailableServices: availableServices,
	}, nil
}

func NewZoneIngressOverviews(zoneIngresses ZoneIngressResourceList, insights ZoneIngressInsightResourceList) ZoneIngressOverviewResourceList {
	insightsByKey := map[model.ResourceKey]*ZoneIngressInsightResource{}
	for _, insight := range insights.Items {
		insightsByKey[model.MetaToResourceKey(insight.Meta)] = insight
	}

	var items []*ZoneIngressOverviewResource
	for _, zoneIngress := range zoneIngresses.Items {
		overview := ZoneIngressOverviewResource{
			Meta: zoneIngress.Meta,
			Spec: &mesh_proto.ZoneIngressOverview{
				ZoneIngress:        zoneIngress.Spec,
				ZoneIngressInsight: nil,
			},
		}
		insight, exists := insightsByKey[model.MetaToResourceKey(overview.Meta)]
		if exists {
			overview.Spec.ZoneIngressInsight = insight.Spec
		}
		items = append(items, &overview)
	}
	return ZoneIngressOverviewResourceList{
		Pagination: zoneIngresses.Pagination,
		Items:      items,
	}
}
