package mesh

import (
	"net"
	"strconv"

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

func (r *ZoneIngressResource) AdminAddress(defaultAdminPort uint32) string {
	if r == nil {
		return ""
	}
	ip := r.Spec.GetNetworking().GetAddress()
	adminPort := r.Spec.GetNetworking().GetAdmin().GetPort()
	if adminPort == 0 {
		adminPort = defaultAdminPort
	}
	return net.JoinHostPort(ip, strconv.FormatUint(uint64(adminPort), 10))
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
