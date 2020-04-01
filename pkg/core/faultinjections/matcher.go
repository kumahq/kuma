package faultinjections

import (
	"context"

	"github.com/pkg/errors"

	core_xds "github.com/Kong/kuma/pkg/core/xds"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/policy"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

type FaultInjectionMatcher struct {
	ResourceManager manager.ReadOnlyResourceManager
}

// policyAdapter allows passing ConnectionPolicy as a DataplanePolicy and return destination Selectors as
// dataplane's Selectors. Source Selectors will be handle on the envoy side with headers matcher regex.
type policyAdapter struct {
	policy.ConnectionPolicy
}

func (p *policyAdapter) Selectors() []*mesh_proto.Selector {
	return p.ConnectionPolicy.Destinations()
}

type list []*mesh_core.FaultInjectionResource

func (l list) asDataplanePolicy() (dataplanePolicies []policy.DataplanePolicy) {
	for _, faultInjection := range l {
		dataplanePolicies = append(dataplanePolicies, &policyAdapter{faultInjection})
	}
	return
}

func (f *FaultInjectionMatcher) Match(ctx context.Context, dataplane *mesh_core.DataplaneResource) (core_xds.FaultInjectionMap, error) {
	faultInjections := &mesh_core.FaultInjectionResourceList{}
	if err := f.ResourceManager.List(ctx, faultInjections, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, errors.Wrap(err, "could not retrieve fault injections")
	}

	matchedFaultInjections := make(core_xds.FaultInjectionMap)
	ifaces, err := dataplane.Spec.GetNetworking().GetInboundInterfaces()
	if err != nil {
		return nil, err
	}

	for i, inbound := range dataplane.Spec.GetNetworking().GetInbound() {
		bestMatch := policy.SelectDataplanePolicy(inbound, list(faultInjections.Items).asDataplanePolicy())
		if bestMatch == nil {
			continue
		}

		matchedFaultInjections[ifaces[i]] = &bestMatch.(policy.ConnectionPolicy).(*policyAdapter).
			ConnectionPolicy.(*mesh_core.FaultInjectionResource).Spec
	}

	return matchedFaultInjections, nil
}
