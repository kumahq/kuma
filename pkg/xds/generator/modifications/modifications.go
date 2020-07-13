package modifications

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	model "github.com/kumahq/kuma/pkg/core/xds"
)

// GeneratedByProxyTemplateModifications is a marker to indicate by which ProxyGenerator resources were generated.
const GeneratedByProxyTemplateModifications = "proxy-template-modifications"

func Apply(resources *model.ResourceSet, modifications []*mesh_proto.ProxyTemplate_Modifications) error {
	for _, modification := range modifications {
		switch modification.Type.(type) {
		case *mesh_proto.ProxyTemplate_Modifications_Cluster_:
			if err := applyClusterModification(resources, modification.GetCluster()); err != nil {
				return err
			}
		case *mesh_proto.ProxyTemplate_Modifications_NetworkFilter_:
			if err := applyNetworkFilterModification(resources, modification.GetNetworkFilter()); err != nil {
				return err
			}
		}
	}
	return nil
}
