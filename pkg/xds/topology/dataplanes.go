package topology

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/resources/store"

	"github.com/go-logr/logr"

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

// GetExternalServices returns list of ExternalServices in provided Mesh
func GetExternalServices(log logr.Logger, ctx context.Context, rm manager.ReadOnlyResourceManager, mesh string) (*core_mesh.ExternalServiceResourceList, error) {
	externalServices := &core_mesh.ExternalServiceResourceList{}
	if err := rm.List(ctx, externalServices, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}

	return externalServices, nil
}

// ResolveAddress resolves 'dataplane.networking.address' if it has DNS name in it. This is a crucial feature for
// some environments specifically AWS ECS. Dataplane resource has to be created before running Kuma DP, but IP address
// will be assigned only after container's start. Envoy EDS doesn't support DNS names, that's why Kuma CP resolves
// addresses before sending resources to the proxy.
func ResolveAddress(lookupIPFunc lookup.LookupIPFunc, dataplane *core_mesh.DataplaneResource) error {
	ips, err := lookupIPFunc(dataplane.Spec.Networking.Address)
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		return errors.Errorf("can't resolve address %v", dataplane.Spec.Networking.Address)
	}
	dataplane.Spec.Networking.Address = ips[0].String()
	return nil
}

func ResolveAddresses(log logr.Logger, lookupIPFunc lookup.LookupIPFunc, dataplanes []*core_mesh.DataplaneResource) []*core_mesh.DataplaneResource {
	rv := []*core_mesh.DataplaneResource{}
	for _, d := range dataplanes {
		if err := ResolveAddress(lookupIPFunc, d); err != nil {
			log.Error(err, "failed to resolve dataplane's domain name, skipping dataplane")
			continue
		}
		rv = append(rv, d)
	}
	return rv
}
