package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func RegisterK8SGatewayTypes() {
	SchemeBuilder.Register(&Gateway{}, &GatewayList{})
	// RegisterObjectTypeIfAbsent is used because it's not deterministic in testing that RegisterGatewayTypes is called only once.
	registry.RegisterObjectTypeIfAbsent(&mesh_proto.Gateway{}, &Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "Gateway",
		},
	})
	registry.RegisterListTypeIfAbsent(&mesh_proto.Gateway{}, &GatewayList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "GatewayList",
		},
	})

	SchemeBuilder.Register(&GatewayRoute{}, &GatewayRouteList{})
	registry.RegisterObjectTypeIfAbsent(&mesh_proto.GatewayRoute{}, &GatewayRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "GatewayRoute",
		},
	})
	registry.RegisterListTypeIfAbsent(&mesh_proto.GatewayRoute{}, &GatewayRouteList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "GatewayRouteList",
		},
	})
	SchemeBuilder.Register(&GatewayInstance{}, &GatewayInstanceList{})
}
