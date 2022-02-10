package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func RegisterK8SGatewayTypes() {
	SchemeBuilder.Register(&MeshGateway{}, &MeshGatewayList{})
	// RegisterObjectTypeIfAbsent is used because it's not deterministic in testing that RegisterMeshGatewayTypes is called only once.
	registry.RegisterObjectTypeIfAbsent(&mesh_proto.MeshGateway{}, &MeshGateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "MeshGateway",
		},
	})
	registry.RegisterListTypeIfAbsent(&mesh_proto.MeshGateway{}, &MeshGatewayList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "MeshGatewayList",
		},
	})

	SchemeBuilder.Register(&MeshGatewayRoute{}, &MeshGatewayRouteList{})
	registry.RegisterObjectTypeIfAbsent(&mesh_proto.MeshGatewayRoute{}, &MeshGatewayRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "MeshGatewayRoute",
		},
	})
	registry.RegisterListTypeIfAbsent(&mesh_proto.MeshGatewayRoute{}, &MeshGatewayRouteList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "MeshGatewayRouteList",
		},
	})
	SchemeBuilder.Register(&MeshGatewayInstance{}, &MeshGatewayInstanceList{})
}
