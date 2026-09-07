package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	bootstrap_k8s "github.com/kumahq/kuma/v3/pkg/plugins/bootstrap/k8s"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
)

func TestRemoveGatewayClassFinalizers(t *testing.T) {
	scheme, err := bootstrap_k8s.NewScheme()
	require.NoError(t, err)

	kumaClass := &gatewayapi.GatewayClass{
		Name:       "kuma",
		Finalizers: []string{gatewayapi_v1.GatewayClassFinalizerGatewaysExist},
		Spec: gatewayapi.GatewayClassSpec{
			ControllerName: common.ControllerName,
		},
	}
	otherClass := &gatewayapi.GatewayClass{
		Name:       "other",
		Finalizers: []string{gatewayapi_v1.GatewayClassFinalizerGatewaysExist},
		Spec: gatewayapi.GatewayClassSpec{
			ControllerName: "example.com/other",
		},
	}
	client := kube_client_fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(kumaClass, otherClass).
		Build()

	require.NoError(t, removeGatewayClassFinalizers(context.Background(), client, client))

	var updatedKumaClass gatewayapi.GatewayClass
	require.NoError(t, client.Get(context.Background(), kube_client.ObjectKeyFromObject(kumaClass), &updatedKumaClass))
	require.NotContains(t, updatedKumaClass.Finalizers, gatewayapi_v1.GatewayClassFinalizerGatewaysExist)

	var updatedOtherClass gatewayapi.GatewayClass
	require.NoError(t, client.Get(context.Background(), kube_client.ObjectKeyFromObject(otherClass), &updatedOtherClass))
	require.Contains(t, updatedOtherClass.Finalizers, gatewayapi_v1.GatewayClassFinalizerGatewaysExist)
}

func TestGRPCRouteCRDsPresentWithV1GRPCRoute(t *testing.T) {
	mapper := newGatewayRESTMapper()
	mapper.Add(gatewaySchemaGroupVersion(gatewayapi_v1.GroupVersion.Version).WithKind("GRPCRoute"), kube_apimeta.RESTScopeNamespace)
	mapper.Add(gatewaySchemaGroupVersion(gatewayapi.GroupVersion.Version).WithKind("ReferenceGrant"), kube_apimeta.RESTScopeNamespace)

	ok, missing := gatewayAPIRESTMappingsPresent(mapper, requiredGRPCRouteCRDs)

	require.True(t, ok)
	require.Empty(t, missing)
}

func TestGRPCRouteCRDsMissingWithOnlyV1Beta1GRPCRoute(t *testing.T) {
	mapper := newGatewayRESTMapper()
	mapper.Add(gatewaySchemaGroupVersion(gatewayapi.GroupVersion.Version).WithKind("GRPCRoute"), kube_apimeta.RESTScopeNamespace)
	mapper.Add(gatewaySchemaGroupVersion(gatewayapi.GroupVersion.Version).WithKind("ReferenceGrant"), kube_apimeta.RESTScopeNamespace)

	ok, missing := gatewayAPIRESTMappingsPresent(mapper, requiredGRPCRouteCRDs)

	require.False(t, ok)
	require.Equal(t, []string{"GRPCRoute.gateway.networking.k8s.io/v1"}, missing)
}

func newGatewayRESTMapper() *kube_apimeta.DefaultRESTMapper {
	return kube_apimeta.NewDefaultRESTMapper([]schema.GroupVersion{
		gatewaySchemaGroupVersion(gatewayapi_v1.GroupVersion.Version),
		gatewaySchemaGroupVersion(gatewayapi.GroupVersion.Version),
	})
}

func gatewaySchemaGroupVersion(version string) schema.GroupVersion {
	return schema.GroupVersion{
		Group:   gatewayapi.GroupVersion.Group,
		Version: version,
	}
}
