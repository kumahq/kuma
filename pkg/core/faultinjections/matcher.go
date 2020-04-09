package faultinjections

import (
	"context"

	"github.com/pkg/errors"

	core_xds "github.com/Kong/kuma/pkg/core/xds"

	"github.com/Kong/kuma/pkg/core/policy"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

type FaultInjectionMatcher struct {
	ResourceManager manager.ReadOnlyResourceManager
}

func (f *FaultInjectionMatcher) Match(ctx context.Context, dataplane *mesh_core.DataplaneResource) (core_xds.FaultInjectionMap, error) {
	faultInjections := &mesh_core.FaultInjectionResourceList{}
	if err := f.ResourceManager.List(ctx, faultInjections, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, errors.Wrap(err, "could not retrieve fault injections")
	}

	policies := make([]policy.ConnectionPolicy, len(faultInjections.Items))
	for i, faultInjection := range faultInjections.Items {
		policies[i] = faultInjection
	}

	policyMap, err := policy.SelectInboundConnectionPolicies(dataplane, policies)
	if err != nil {
		return nil, err
	}

	result := core_xds.FaultInjectionMap{}
	for inbound, connectionPolicy := range policyMap {
		result[inbound] = &connectionPolicy.(*mesh_core.FaultInjectionResource).Spec
	}
	return result, nil
}
