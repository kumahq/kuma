package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/jsonpatch"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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

			util_proto.Merge(cluster.Resource, clusterMod)
		}
	}

	return nil
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
		Origin:   Origin,
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
	if c.Match.Origin != nil && *c.Match.Origin != cluster.Origin {
		return false
	}
	return true
}
