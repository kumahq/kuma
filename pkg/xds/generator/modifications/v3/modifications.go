package v3

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

type modificator interface {
	apply(*core_xds.ResourceSet) error
}

func Apply(resources *core_xds.ResourceSet, modifications []*mesh_proto.ProxyTemplate_Modifications) error {
	for i, modification := range modifications {
		var modificator modificator
		switch modification.Type.(type) {
		case *mesh_proto.ProxyTemplate_Modifications_Cluster_:
			mod := clusterModificator(*modification.GetCluster())
			modificator = &mod
		case *mesh_proto.ProxyTemplate_Modifications_Listener_:
			mod := listenerModificator(*modification.GetListener())
			modificator = &mod
		case *mesh_proto.ProxyTemplate_Modifications_NetworkFilter_:
			mod := networkFilterModificator(*modification.GetNetworkFilter())
			modificator = &mod
		case *mesh_proto.ProxyTemplate_Modifications_HttpFilter_:
			mod := httpFilterModificator(*modification.GetHttpFilter())
			modificator = &mod
		case *mesh_proto.ProxyTemplate_Modifications_VirtualHost_:
			mod := virtualHostModificator(*modification.GetVirtualHost())
			modificator = &mod
		default:
			return errors.Errorf("invalid modification type %T", modification.Type)
		}
		if err := modificator.apply(resources); err != nil {
			return errors.Wrapf(err, "could not apply %d modification", i)
		}
	}
	return nil
}
