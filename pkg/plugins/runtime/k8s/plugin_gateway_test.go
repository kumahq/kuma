package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	bootstrap_k8s "github.com/kumahq/kuma/v3/pkg/plugins/bootstrap/k8s"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
)

func TestCleanupLegacyGatewayClassFinalizers(t *testing.T) {
	scheme, err := bootstrap_k8s.NewScheme()
	require.NoError(t, err)

	kumaClass := &gatewayapi.GatewayClass{
		ObjectMeta: kube_meta.ObjectMeta{
			Name:       "kuma",
			Finalizers: []string{gatewayapi_v1.GatewayClassFinalizerGatewaysExist},
		},
		Spec: gatewayapi.GatewayClassSpec{
			ControllerName: common.ControllerName,
		},
	}
	otherClass := &gatewayapi.GatewayClass{
		ObjectMeta: kube_meta.ObjectMeta{
			Name:       "other",
			Finalizers: []string{gatewayapi_v1.GatewayClassFinalizerGatewaysExist},
		},
		Spec: gatewayapi.GatewayClassSpec{
			ControllerName: "example.com/other",
		},
	}
	client := kube_client_fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(kumaClass, otherClass).
		Build()

	require.NoError(t, cleanupLegacyGatewayClassFinalizers(context.Background(), client))

	var updatedKumaClass gatewayapi.GatewayClass
	require.NoError(t, client.Get(context.Background(), kube_client.ObjectKeyFromObject(kumaClass), &updatedKumaClass))
	require.NotContains(t, updatedKumaClass.Finalizers, gatewayapi_v1.GatewayClassFinalizerGatewaysExist)

	var updatedOtherClass gatewayapi.GatewayClass
	require.NoError(t, client.Get(context.Background(), kube_client.ObjectKeyFromObject(otherClass), &updatedOtherClass))
	require.Contains(t, updatedOtherClass.Finalizers, gatewayapi_v1.GatewayClassFinalizerGatewaysExist)
}
