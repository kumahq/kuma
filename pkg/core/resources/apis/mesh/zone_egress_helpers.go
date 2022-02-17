package mesh

import (
	"net"
	"strconv"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func (r *ZoneEgressResource) UsesInboundInterface(address net.IP, port uint32) bool {
	if r == nil {
		return false
	}

	if port == r.Spec.GetNetworking().GetPort() && overlap(address, net.ParseIP(r.Spec.GetNetworking().GetAddress())) {
		return true
	}

	return false
}

func (r *ZoneEgressResource) IsIPv6() bool {
	if r == nil {
		return false
	}

	ip := net.ParseIP(r.Spec.GetNetworking().GetAddress())
	if ip == nil {
		return false
	}

	return ip.To4() == nil
}

func (r *ZoneEgressResource) AdminAddress(defaultAdminPort uint32) string {
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

func NewZoneEgressOverviews(zoneEgresses ZoneEgressResourceList, insights ZoneEgressInsightResourceList) ZoneEgressOverviewResourceList {
	insightsByKey := map[model.ResourceKey]*ZoneEgressInsightResource{}
	for _, insight := range insights.Items {
		insightsByKey[model.MetaToResourceKey(insight.Meta)] = insight
	}

	var items []*ZoneEgressOverviewResource
	for _, zoneEgress := range zoneEgresses.Items {
		overview := ZoneEgressOverviewResource{
			Meta: zoneEgress.Meta,
			Spec: &mesh_proto.ZoneEgressOverview{
				ZoneEgress:        zoneEgress.Spec,
				ZoneEgressInsight: nil,
			},
		}
		insight, exists := insightsByKey[model.MetaToResourceKey(overview.Meta)]
		if exists {
			overview.Spec.ZoneEgressInsight = insight.Spec
		}
		items = append(items, &overview)
	}
	return ZoneEgressOverviewResourceList{
		Pagination: zoneEgresses.Pagination,
		Items:      items,
	}
}
