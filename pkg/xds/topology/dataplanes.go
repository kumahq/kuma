package topology

import (
	"context"
	"net"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/resources/manager"

	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

// GetDataplanes returns list of Dataplane in provided Mesh and Ingresses (which are cluster-scoped, not mesh-scoped)
func GetDataplanes(log logr.Logger, ctx context.Context, rm manager.ReadOnlyResourceManager, lookupIPFunc lookup.LookupIPFunc, mesh string) (*core_mesh.DataplaneResourceList, error) {
	dataplanes := &core_mesh.DataplaneResourceList{}
	if err := rm.List(ctx, dataplanes); err != nil {
		return nil, err
	}
	dataplanes.Items = ResolveAddresses(log, lookupIPFunc, dataplanes.Items)
	filteredDataplanes := &core_mesh.DataplaneResourceList{}
	for _, d := range dataplanes.Items {
		if d.GetMeta().GetMesh() == mesh || d.Spec.IsIngress() {
			_ = filteredDataplanes.AddItem(d)
		}
	}

	return filteredDataplanes, nil
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
	if dataplane.Spec.Networking.AdvertiseAddress != "" {
		if aips, err = lookupIPFunc(dataplane.Spec.Networking.AdvertiseAddress); err != nil {
			return nil, err
		}
		if len(aips) == 0 {
			return nil, errors.Errorf("can't resolve address %v", dataplane.Spec.Networking.AdvertiseAddress)
		}
		if dataplane.Spec.Networking.AdvertiseAddress != aips[0].String() {
			update_aip = true
		}
	}

	if update_ip || update_aip { // only if we resolve any address, in most cases this is IP not a hostname
		dpSpec := proto.Clone(dataplane.Spec).(*mesh_proto.Dataplane)
		if update_ip {
			dpSpec.Networking.Address = ips[0].String()
		}
		if update_aip {
			dpSpec.Networking.AdvertiseAddress = aips[0].String()
		}
		return &core_mesh.DataplaneResource{
			Meta: dataplane.Meta,
			Spec: dpSpec,
		}, nil
	}
	return dataplane, nil
}

func ResolveIngressPublicAddress(lookupIPFunc lookup.LookupIPFunc, dataplane *core_mesh.DataplaneResource) (*core_mesh.DataplaneResource, error) {
	if dataplane.Spec.Networking.Ingress.PublicAddress == "" { // Ingress may not have public address yet.
		return dataplane, nil
	}
	ips, err := lookupIPFunc(dataplane.Spec.Networking.Ingress.PublicAddress)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, errors.Errorf("can't resolve address %v", dataplane.Spec.Networking.Ingress.PublicAddress)
	}
	if dataplane.Spec.Networking.Ingress.PublicAddress != ips[0].String() { // only if we resolve any address, in most cases this is IP not a hostname
		dpSpec := proto.Clone(dataplane.Spec).(*mesh_proto.Dataplane)
		dpSpec.Networking.Ingress.PublicAddress = ips[0].String()
		return &core_mesh.DataplaneResource{
			Meta: dataplane.Meta,
			Spec: dpSpec,
		}, nil
	}
	return dataplane, nil
}

func ResolveAddresses(log logr.Logger, lookupIPFunc lookup.LookupIPFunc, dataplanes []*core_mesh.DataplaneResource) []*core_mesh.DataplaneResource {
	rv := []*core_mesh.DataplaneResource{}
	for _, d := range dataplanes {
		if d.Spec.IsIngress() {
			dp, err := ResolveIngressPublicAddress(lookupIPFunc, d)
			if err != nil {
				log.Error(err, "failed to resolve ingress's public name, skipping dataplane")
				continue
			}
			rv = append(rv, dp)
		} else {
			dp, err := ResolveAddress(lookupIPFunc, d)
			if err != nil {
				log.Error(err, "failed to resolve dataplane's domain name, skipping dataplane")
				continue
			}
			rv = append(rv, dp)
		}
	}
	return rv
}
