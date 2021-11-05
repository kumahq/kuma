//go:build gateway
// +build gateway

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func init() {
	SchemeBuilder.Register(&Gateway{}, &GatewayList{})
	registry.RegisterObjectType(&mesh_proto.Gateway{}, &Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "Gateway",
		},
	})
	registry.RegisterListType(&mesh_proto.Gateway{}, &GatewayList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "GatewayList",
		},
	})
}
