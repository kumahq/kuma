package mesh

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

func ProxyTypeFromResourceType(t core_model.ResourceType) (mesh_proto.ProxyType, error) {
	switch t {
	case DataplaneType:
		return mesh_proto.DataplaneProxyType, nil
	case ZoneIngressType:
		return mesh_proto.IngressProxyType, nil
	case ZoneEgressType:
		return mesh_proto.EgressProxyType, nil
	default:
		return "", errors.Errorf("%s does not have a corresponding proxy type", t)
	}
}

func ResourceTypeDescriptorFromProxyType(proxyType mesh_proto.ProxyType) (core_model.ResourceTypeDescriptor, error) {
	switch proxyType {
	case mesh_proto.DataplaneProxyType:
		return DataplaneResourceTypeDescriptor, nil
	case mesh_proto.IngressProxyType:
		return ZoneIngressResourceTypeDescriptor, nil
	case mesh_proto.EgressProxyType:
		return ZoneEgressResourceTypeDescriptor, nil
	default:
		return core_model.ResourceTypeDescriptor{}, errors.Errorf("%s does not have a corresponding resource type", proxyType)
	}
}
