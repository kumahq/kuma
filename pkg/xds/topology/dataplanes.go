package topology

import (
	"context"
	"net"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
)

func GetZoneIngresses(log logr.Logger, ctx context.Context, rm manager.ReadOnlyResourceManager, lookupIPFunc lookup.LookupIPFunc) (*core_mesh.ZoneIngressResourceList, error) {
	zoneIngresses := &core_mesh.ZoneIngressResourceList{}
	if err := rm.List(ctx, zoneIngresses); err != nil {
		return nil, err
	}
	zoneIngresses.Items = ResolveZoneIngressAddresses(log, lookupIPFunc, zoneIngresses.Items)
	return zoneIngresses, nil
}

// ResolveAddress resolves 'dataplane.networking.address' if it has DNS name in it. This is a crucial feature for
// some environments specifically AWS ECS. Dataplane resource has to be created before running Kuma DP, but IP address
// will be assigned only after container's start. Envoy EDS doesn't support DNS names, that's why Kuma CP resolves
// addresses before sending resources to the proxy.
func ResolveAddress(lookupIPFunc lookup.LookupIPFunc, dataplane *core_mesh.DataplaneResource) (*core_mesh.DataplaneResource, error) {
	var ips, aips []net.IP
	var err error
	var update_ip, update_aip bool = false, false
	if ips, err = lookupIPFunc(dataplane.Spec.Networking.Address); err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, errors.Errorf("can't resolve address %v", dataplane.Spec.Networking.Address)
	}
	if dataplane.Spec.Networking.Address != ips[0].String() {
		update_ip = true
	}
	if dataplane.Spec.Networking.AdvertisedAddress != "" {
		if aips, err = lookupIPFunc(dataplane.Spec.Networking.AdvertisedAddress); err != nil {
			return nil, err
		}
		if len(aips) == 0 {
			return nil, errors.Errorf("can't resolve address %v", dataplane.Spec.Networking.AdvertisedAddress)
		}
		if dataplane.Spec.Networking.AdvertisedAddress != aips[0].String() {
			update_aip = true
		}
	}

	if update_ip || update_aip { // only if we resolve any address, in most cases this is IP not a hostname
		dpSpec := proto.Clone(dataplane.Spec).(*mesh_proto.Dataplane)
		if update_ip {
			dpSpec.Networking.Address = ips[0].String()
		}
		if update_aip {
			dpSpec.Networking.AdvertisedAddress = aips[0].String()
		}
		return &core_mesh.DataplaneResource{
			Meta: dataplane.Meta,
			Spec: dpSpec,
		}, nil
	}
	return dataplane, nil
}

func ResolveZoneIngressPublicAddress(lookupIPFunc lookup.LookupIPFunc, zoneIngress *core_mesh.ZoneIngressResource) (*core_mesh.ZoneIngressResource, error) {
	if zoneIngress.Spec.GetNetworking().GetAdvertisedAddress() == "" { // Ingress may not have public address yet.
		return zoneIngress, nil
	}
	ips, err := lookupIPFunc(zoneIngress.Spec.GetNetworking().GetAdvertisedAddress())
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, errors.Errorf("can't resolve address %v", zoneIngress.Spec.GetNetworking().GetAdvertisedAddress())
	}
	if zoneIngress.Spec.GetNetworking().GetAdvertisedAddress() != ips[0].String() { // only if we resolve any address, in most cases this is IP not a hostname
		ziSpec := proto.Clone(zoneIngress.Spec).(*mesh_proto.ZoneIngress)
		ziSpec.Networking.AdvertisedAddress = ips[0].String()
		return &core_mesh.ZoneIngressResource{
			Meta: zoneIngress.Meta,
			Spec: ziSpec,
		}, nil
	}
	return zoneIngress, nil
}

func ResolveAddresses(log logr.Logger, lookupIPFunc lookup.LookupIPFunc, dataplanes []*core_mesh.DataplaneResource) []*core_mesh.DataplaneResource {
	rv := []*core_mesh.DataplaneResource{}
	for _, d := range dataplanes {
		dp, err := ResolveAddress(lookupIPFunc, d)
		if err != nil {
			log.Error(err, "failed to resolve dataplane's domain name, skipping dataplane")
			continue
		}
		rv = append(rv, dp)
	}
	return rv
}

func ResolveZoneIngressAddresses(log logr.Logger, lookupIPFunc lookup.LookupIPFunc, zoneIngresses []*core_mesh.ZoneIngressResource) []*core_mesh.ZoneIngressResource {
	rv := []*core_mesh.ZoneIngressResource{}
	for _, zi := range zoneIngresses {
		resolvedZi, err := ResolveZoneIngressPublicAddress(lookupIPFunc, zi)
		if err != nil {
			log.Error(err, "failed to resolve ingress's public name, skipping dataplane")
			continue
		}
		rv = append(rv, resolvedZi)
	}
	return rv
}
