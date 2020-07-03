package modifications

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	model "github.com/Kong/kuma/pkg/core/xds"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func applyClusterModification(resources *model.ResourceSet, modification *mesh_proto.ProxyTemplate_Modifications_Cluster) error {
	clusterMod := &envoy_api.Cluster{}
	if err := util_proto.FromYAML([]byte(modification.Value), clusterMod); err != nil {
		return err
	}
	switch modification.Operation {
	case "add":
		resources.AddNamed(clusterMod)
	case "remove":
		for name, resource := range resources.Resources(envoy_resource.ClusterType) {
			if clusterMatches(resource, modification.Match) {
				resources.Remove(envoy_resource.ClusterType, name)
			}
		}
	case "patch":
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
	// todo support side cluster.Side == "inbound"
	return false
}
