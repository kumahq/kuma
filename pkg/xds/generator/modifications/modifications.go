package modifications

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	modifications_v3 "github.com/kumahq/kuma/pkg/xds/generator/modifications/v3"
)

func Apply(resources *core_xds.ResourceSet, modifications []*mesh_proto.ProxyTemplate_Modifications, apiVersion core_xds.APIVersion) error {
	switch apiVersion {
	case envoy_common.APIV3:
		return modifications_v3.Apply(resources, modifications)
	default:
		return errors.Errorf("unknown API version %s", apiVersion)
	}
}
