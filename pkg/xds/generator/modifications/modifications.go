package modifications

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	model "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/pkg/errors"
)

// OriginProxyTemplateModifications is a marker to indicate by which ProxyGenerator resources were generated.
const OriginProxyTemplateModifications = "proxy-template-modifications"

func Apply(resources *model.ResourceSet, modifications []*mesh_proto.ProxyTemplate_Modifications) error {
	for _, modification := range modifications {
		switch modification.Type.(type) {
		case *mesh_proto.ProxyTemplate_Modifications_Cluster_:
			if err := applyClusterModification(resources, modification.GetCluster()); err != nil {
				return errors.Wrap(err, "could not apply cluster modification")
			}
		case *mesh_proto.ProxyTemplate_Modifications_Listener_:
			if err := applyListenerModification(resources, modification.GetListener()); err != nil {
				return errors.Wrap(err, "could not apply listener modification")
			}
		case *mesh_proto.ProxyTemplate_Modifications_NetworkFilter_:
			if err := applyNetworkFilterModification(resources, modification.GetNetworkFilter()); err != nil {
				return errors.Wrap(err, "could not apply network filter modification")
			}
		case *mesh_proto.ProxyTemplate_Modifications_HttpFilter_:
			if err := applyHTTPFilterModification(resources, modification.GetHttpFilter()); err != nil {
				return errors.Wrap(err, "could not apply HTTP filter modification")
			}
		}
	}
	return nil
}
