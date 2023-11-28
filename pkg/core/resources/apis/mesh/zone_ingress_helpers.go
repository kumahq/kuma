package mesh

import (
	"hash/fnv"
	"net"
	"strconv"

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

func (r *ZoneIngressResource) Hash() []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(model.HashMeta(r))
	_, _ = hasher.Write([]byte(r.Spec.GetNetworking().GetAddress()))
	_, _ = hasher.Write([]byte(r.Spec.GetNetworking().GetAdvertisedAddress()))
	return hasher.Sum(nil)
}
