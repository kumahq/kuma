package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/jsonpatch"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

type clusterModificator api.ClusterMod

func (c *clusterModificator) apply(resources *core_xds.ResourceSet) error {
	clusterMod := &envoy_cluster.Cluster{}
	if c.Value != nil {
		if err := util_proto.FromYAML([]byte(*c.Value), clusterMod); err != nil {
			return err
		}
	}

	switch c.Operation {
	case api.ModOpAdd:
		c.add(resources, clusterMod)
	case api.ModOpRemove:
		c.remove(resources)
	case api.ModOpPatch:
		return c.patch(resources, clusterMod)
	default:
		return errors.Errorf("invalid operation: %s", c.Operation)
	}
	return nil
}

func (c *clusterModificator) patch(resources *core_xds.ResourceSet, clusterMod *envoy_cluster.Cluster) error {
	for _, cluster := range resources.Resources(envoy_resource.ClusterType) {
		if c.clusterMatches(cluster) {
			if len(pointer.Deref(c.JsonPatches)) > 0 {
				if err := jsonpatch.MergeJsonPatch(cluster.Resource, pointer.Deref(c.JsonPatches)); err != nil {
					return err
				}

				continue
			}

			mod := clusterMod
			if target, ok := cluster.Resource.(*envoy_cluster.Cluster); ok && clusterMod.GetCircuitBreakers() != nil {
				// copied because the modification is shared by every cluster it matches
				mod = proto.Clone(clusterMod).(*envoy_cluster.Cluster)
				mergeCircuitBreakers(target, mod.CircuitBreakers)
			}

			util_proto.Merge(cluster.Resource, mod)
		}
	}

	return nil
}

// mergeCircuitBreakers merges the patched thresholds into the ones the cluster
// already carries and takes them out of the patch, so that the merge that follows
// doesn't append them a second time.
//
// Envoy reads a single threshold per routing priority and ignores every later
// entry for that priority, while proto merge appends to repeated fields. Since
// every generated cluster already carries a DEFAULT threshold, an appended patch
// would end up second in the list and never take effect.
func mergeCircuitBreakers(cluster *envoy_cluster.Cluster, patch *envoy_cluster.CircuitBreakers) {
	if cluster.CircuitBreakers == nil {
		cluster.CircuitBreakers = &envoy_cluster.CircuitBreakers{}
	}
	breakers := cluster.CircuitBreakers

	breakers.Thresholds = mergeThresholds(breakers.Thresholds, patch.Thresholds)
	breakers.PerHostThresholds = mergeThresholds(breakers.PerHostThresholds, patch.PerHostThresholds)
	patch.Thresholds, patch.PerHostThresholds = nil, nil
}

// mergeThresholds merges every patched threshold into the entries it applies to.
// A patch that names a priority updates the entry for that priority, and a patch
// that leaves the priority out updates all of them, since limits written without
// a priority are meant for the whole cluster. Priorities the cluster doesn't have
// yet are appended.
func mergeThresholds(
	thresholds []*envoy_cluster.CircuitBreakers_Thresholds,
	patches []*envoy_cluster.CircuitBreakers_Thresholds,
) []*envoy_cluster.CircuitBreakers_Thresholds {
	for _, patch := range patches {
		applied := false

		for _, threshold := range thresholds {
			if patch.GetPriority() == envoy_core.RoutingPriority_DEFAULT ||
				patch.GetPriority() == threshold.GetPriority() {
				util_proto.Merge(threshold, patch)
				applied = true
			}
		}

		if !applied {
			thresholds = append(thresholds, patch)
		}
	}

	return thresholds
}

func (c *clusterModificator) remove(resources *core_xds.ResourceSet) {
	for name, resource := range resources.Resources(envoy_resource.ClusterType) {
		if c.clusterMatches(resource) {
			resources.Remove(envoy_resource.ClusterType, name)
		}
	}
}

func (c *clusterModificator) add(resources *core_xds.ResourceSet, clusterMod *envoy_cluster.Cluster) *core_xds.ResourceSet {
	return resources.Add(&core_xds.Resource{
		Name:     clusterMod.Name,
		Origin:   metadata.OriginProxyTemplateModifications,
		Resource: clusterMod,
	})
}

func (c *clusterModificator) clusterMatches(cluster *core_xds.Resource) bool {
	if c.Match == nil {
		return true
	}
	if c.Match.Name != nil && *c.Match.Name != cluster.Name {
		return false
	}
	if c.Match.Origin != nil && *c.Match.Origin != string(cluster.Origin) {
		return false
	}
	return true
}
