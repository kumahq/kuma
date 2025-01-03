package mesh

import (
	"hash/fnv"
	"net"
	"strconv"

	"github.com/kumahq/kuma/api/common/v1alpha1"
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

func (r *ZoneEgressResource) Hash() []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(model.HashMeta(r))
	_, _ = hasher.Write([]byte(r.Spec.GetNetworking().GetAddress()))
	return hasher.Sum(nil)
}

func (r *ZoneEgressResource) IsRemoteEgress(localZone string) bool {
	return r.Spec.GetZone() != "" && r.Spec.GetZone() != localZone
}

func (r *ZoneEgressResource) GetProxyType() v1alpha1.TargetRefProxyType {
	return v1alpha1.ZoneEgress
}
