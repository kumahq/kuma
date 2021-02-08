package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type clusterModificator mesh_proto.ProxyTemplate_Modifications_Cluster

func (c *clusterModificator) apply(resources *core_xds.ResourceSet) error {
	clusterMod := &envoy_api.Cluster{}
	if err := util_proto.FromYAML([]byte(c.Value), clusterMod); err != nil {
		return err
	}
	switch c.Operation {
	case mesh_proto.OpAdd:
		c.add(resources, clusterMod)
	case mesh_proto.OpRemove:
		c.remove(resources)
	case mesh_proto.OpPatch:
		c.patch(resources, clusterMod)
	default:
		return errors.Errorf("invalid operation: %s", c.Operation)
	}
	return nil
}

func (c *clusterModificator) patch(resources *core_xds.ResourceSet, clusterMod *envoy_api.Cluster) {
	for _, cluster := range resources.Resources(envoy_resource.ClusterType) {
		if c.clusterMatches(cluster) {
			proto.Merge(cluster.Resource, clusterMod)
		}
	}
}

func (c *clusterModificator) remove(resources *core_xds.ResourceSet) {
	for name, resource := range resources.Resources(envoy_resource.ClusterType) {
		if c.clusterMatches(resource) {
			resources.Remove(envoy_resource.ClusterType, name)
		}
	}
}

func (c *clusterModificator) add(resources *core_xds.ResourceSet, clusterMod *envoy_api.Cluster) *core_xds.ResourceSet {
	return resources.Add(&core_xds.Resource{
		Name:     clusterMod.Name,
		Origin:   OriginProxyTemplateModifications,
		Resource: clusterMod,
	})
}

func (c *clusterModificator) clusterMatches(cluster *core_xds.Resource) bool {
	if c.Match.GetName() != "" && c.Match.GetName() != cluster.Name {
		return false
	}
	if c.Match.GetOrigin() != "" && c.Match.GetOrigin() != cluster.Origin {
		return false
	}
	return true
}
