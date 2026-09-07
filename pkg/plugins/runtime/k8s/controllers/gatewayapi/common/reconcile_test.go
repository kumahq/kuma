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

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
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

func TestReconcileLabelledObjectMigratesDesiredObjectWithLegacyOwnerLabel(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	scheme := kube_runtime.NewScheme()
	g.Expect(meshhttproute_k8s.AddToScheme(scheme)).To(Succeed())

	owner := kube_types.NamespacedName{Namespace: "route-ns", Name: "my-route"}
	existing := &meshhttproute_k8s.MeshHTTPRoute{
		Name:      "generated-route",
		Namespace: "route-ns",
		Labels: map[string]string{
			ownerLabel: LegacyOwnerLabelValue(owner),
			"team":     "payments",
		},
		Spec: &meshhttproute_api.MeshHTTPRoute{},
	}

	client := kube_client_fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(existing).
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
			"generated-route": {
				Namespace: "route-ns",
				Spec:      &meshhttproute_api.MeshHTTPRoute{},
			},
		},
		LegacyOwnerLabelValue(owner),
	)
	g.Expect(err).ToNot(HaveOccurred())

	updated := &meshhttproute_k8s.MeshHTTPRoute{}
	g.Expect(client.Get(ctx, kube_types.NamespacedName{Namespace: "route-ns", Name: "generated-route"}, updated)).To(Succeed())
	g.Expect(updated.Labels).To(HaveKeyWithValue(ownerLabel, hashNamespacedName("HTTPRoute", owner)))
	g.Expect(updated.Labels).To(HaveKeyWithValue("team", "payments"))
}

func TestReconcileLabelledObjectDeletesUnneededObjectWithLegacyOwnerLabel(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	scheme := kube_runtime.NewScheme()
	g.Expect(meshhttproute_k8s.AddToScheme(scheme)).To(Succeed())

	owner := kube_types.NamespacedName{Namespace: "route-ns", Name: "my-route"}
	stale := &meshhttproute_k8s.MeshHTTPRoute{
		Name:      "legacy-route",
		Namespace: "kuma-system",
		Labels: map[string]string{
			ownerLabel: LegacyOwnerLabelValue(owner),
		},
	}

	client := kube_client_fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(stale).
		Build()

	err := ReconcileLabelledObject(
		ctx,
		logr.Discard(),
		k8s_registry.Global(),
		client,
		"HTTPRoute",
		owner,
		core_model.NoMesh,
		&meshhttproute_api.MeshHTTPRoute{},
		nil,
		LegacyOwnerLabelValue(owner),
	)
	g.Expect(err).ToNot(HaveOccurred())

	routes := &meshhttproute_k8s.MeshHTTPRouteList{}
	g.Expect(client.List(ctx, routes)).To(Succeed())
	g.Expect(routes.Items).To(BeEmpty())
}
