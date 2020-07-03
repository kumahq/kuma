package modifications

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	model "github.com/Kong/kuma/pkg/core/xds"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

func applyClusterModification(resources *model.ResourceSet, modification *mesh_proto.ProxyTemplate_Modifications_Cluster) error {
	clusterMod := &envoy_api.Cluster{}
	if err := util_proto.FromYAML([]byte(modification.Value), clusterMod); err != nil {
		return err
	}
	switch modification.Operation {
	case mesh_proto.OpAdd:
		resources.Add(&model.Resource{
			Name:        clusterMod.Name,
			GeneratedBy: GeneratedByProxyTemplateModifications,
			Resource:    clusterMod,
		})
	case mesh_proto.OpRemove:
		for name, resource := range resources.Resources(envoy_resource.ClusterType) {
			if clusterMatches(resource, modification.Match) {
				resources.Remove(envoy_resource.ClusterType, name)
			}
		}
	case mesh_proto.OpPatch:
		for _, cluster := range resources.Resources(envoy_resource.ClusterType) {
			if clusterMatches(cluster, modification.Match) {
				proto.Merge(cluster.Resource, clusterMod)
			}
		}
	default:
		return errors.New("invalid operation")
	}
	return nil
}

func clusterMatches(cluster *model.Resource, match *mesh_proto.ProxyTemplate_Modifications_Cluster_Match) bool {
	if match == nil {
		return true
	}
	if match.Name == cluster.Name {
		return true
	}
	if match.Direction == cluster.GeneratedBy {
		return true
	}
	return false
}
