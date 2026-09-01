package common

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_interceptor "sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_k8s "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/k8s/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
)

func TestReconcileLabelledObjectKeepsUnneededObjectWhenReplacementCreateFails(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	scheme := kube_runtime.NewScheme()
	g.Expect(meshhttproute_k8s.AddToScheme(scheme)).To(Succeed())

	owner := kube_types.NamespacedName{Namespace: "route-ns", Name: "my-route"}
	stale := &meshhttproute_k8s.MeshHTTPRoute{
		Name:      "legacy-route",
		Namespace: "kuma-system",
		Labels: map[string]string{
			ownerLabel: hashNamespacedName("HTTPRoute", owner),
		},
	}

	writeErr := errors.New("simulated replacement create failure")
	deleteCalled := false
	client := kube_client_fake.NewClientBuilder().
		WithScheme(scheme).
		WithInterceptorFuncs(kube_interceptor.Funcs{
			Create: func(ctx context.Context, client kube_client.WithWatch, obj kube_client.Object, opts ...kube_client.CreateOption) error {
				if _, ok := obj.(*meshhttproute_k8s.MeshHTTPRoute); ok {
					return writeErr
				}
				return client.Create(ctx, obj, opts...)
			},
			Delete: func(ctx context.Context, client kube_client.WithWatch, obj kube_client.Object, opts ...kube_client.DeleteOption) error {
				deleteCalled = true
				return client.Delete(ctx, obj, opts...)
			},
		}).
		WithObjects(stale).
		Build()

	err := ReconcileLabelledObject(
		ctx,
		logr.Discard(),
		k8s_registry.Global(),
		client,
		"HTTPRoute",
		owner,
		"default",
		&meshhttproute_api.MeshHTTPRoute{},
		map[string]OwnedObject{
			"replacement-route": {
				Namespace: "kuma-system",
				Spec:      &meshhttproute_api.MeshHTTPRoute{},
			},
		},
	)
	g.Expect(err).To(MatchError(ContainSubstring("could not create owned")))
	g.Expect(err).To(MatchError(ContainSubstring(writeErr.Error())))
	g.Expect(deleteCalled).To(BeFalse())

	routes := &meshhttproute_k8s.MeshHTTPRouteList{}
	g.Expect(client.List(ctx, routes)).To(Succeed())
	g.Expect(routes.Items).To(HaveLen(1))
	g.Expect(routes.Items[0].Name).To(Equal("legacy-route"))
}
