package faultinjections

import (
	"context"
	"fmt"
	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/policy"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/pkg/errors"
)

type FaultInjectionMatcher struct {
	ResourceManager manager.ReadOnlyResourceManager
}

type policyAdapter struct {
	policy.ConnectionPolicy
}

func (p *policyAdapter) Selectors() []*v1alpha1.Selector {
	return p.ConnectionPolicy.Destinations()
}

type list []*mesh_core.FaultInjectionResource

func (l list) AsDataplanePolicy() (dataplanePolicies []policy.DataplanePolicy) {
	for _, faultInjection := range l {
		dataplanePolicies = append(dataplanePolicies, &policyAdapter{faultInjection})
	}
	return
}

func (f *FaultInjectionMatcher) Match(ctx context.Context, dataplane *mesh_core.DataplaneResource) (*mesh_core.FaultInjectionResource, error) {
	faultInjections := &mesh_core.FaultInjectionResourceList{}
	fmt.Println(dataplane.GetMeta())
	if err := f.ResourceManager.List(ctx, faultInjections, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, errors.Wrap(err, "could not retrieve fault injections")
	}

	bestMatch := policy.SelectDataplanePolicy(dataplane, list(faultInjections.Items).AsDataplanePolicy())
	if bestMatch == nil {
		return nil, nil
	}

	return bestMatch.(policy.ConnectionPolicy).(*policyAdapter).ConnectionPolicy.(*mesh_core.FaultInjectionResource), nil
}
