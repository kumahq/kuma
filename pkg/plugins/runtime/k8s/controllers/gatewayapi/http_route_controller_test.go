package gatewayapi

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshservice_k8s "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/k8s/v1alpha1"
	bootstrap_k8s "github.com/kumahq/kuma/v3/pkg/plugins/bootstrap/k8s"
	meshhttproute_k8s "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/k8s/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var _ = Describe("HTTPRouteReconciler.Reconcile with a MeshService parentRef", func() {
	const routeNamespace = "kuma-demo"

	meshServiceParentRef := func(name string) gatewayapi.ParentReference {
		group := gatewayapi.Group(meshservice_k8s.GroupVersion.Group)
		kind := gatewayapi.Kind("MeshService")
		ns := gatewayapi.Namespace(routeNamespace)
		return gatewayapi.ParentReference{
			Group:     &group,
			Kind:      &kind,
			Namespace: &ns,
			Name:      gatewayapi.ObjectName(name),
		}
	}

	withSectionName := func(ref gatewayapi.ParentReference, sectionName string) gatewayapi.ParentReference {
		ref.SectionName = pointer.To(gatewayapi.SectionName(sectionName))
		return ref
	}

	newRoute := func(parentRefs ...gatewayapi.ParentReference) *gatewayapi.HTTPRoute {
		return &gatewayapi.HTTPRoute{
			ObjectMeta: kube_meta.ObjectMeta{Name: "my-route", Namespace: routeNamespace},
			Spec: gatewayapi.HTTPRouteSpec{
				CommonRouteSpec: gatewayapi.CommonRouteSpec{ParentRefs: parentRefs},
			},
		}
	}

	var reconciler *HTTPRouteReconciler
	var namespace *kube_core.Namespace
	var newClientBuilder func(objs ...kube_client.Object) kube_client.Client

	BeforeEach(func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		namespace = &kube_core.Namespace{ObjectMeta: kube_meta.ObjectMeta{Name: "kuma-demo"}}

		newClientBuilder = func(objs ...kube_client.Object) kube_client.Client {
			return kube_client_fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&gatewayapi.HTTPRoute{}).
				WithObjects(append([]kube_client.Object{namespace}, objs...)...).
				Build()
		}

		reconciler = &HTTPRouteReconciler{
			Log:             logr.Discard(),
			TypeRegistry:    k8s_registry.Global(),
			SystemNamespace: "kuma-system",
		}
	})

	It("generates a MeshHTTPRoute and reports ResolvedRefs=True when the MeshService exists", func() {
		ms := &meshservice_k8s.MeshService{
			ObjectMeta: kube_meta.ObjectMeta{Name: "backend", Namespace: "kuma-demo"},
			Spec: &meshservice_api.MeshService{
				Ports: []meshservice_api.Port{{Port: 80, Name: pointer.To("http")}},
			},
		}
		route := newRoute(meshServiceParentRef("backend"))

		client := newClientBuilder(ms, route)
		reconciler.Client = client

		_, err := reconciler.Reconcile(context.Background(), kube_ctrl.Request{
			NamespacedName: kube_client.ObjectKeyFromObject(route),
		})
		Expect(err).ToNot(HaveOccurred())

		routes := &meshhttproute_k8s.MeshHTTPRouteList{}
		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(HaveLen(1))
		Expect(routes.Items[0].Name).To(Equal("my-route-kuma-demo-meshservice-backend.kuma-demo"))

		spec := routes.Items[0].Spec
		Expect(spec).ToNot(BeNil())
		Expect(*spec.To).To(HaveLen(1))
		Expect(*(*spec.To)[0].TargetRef.SectionName).To(Equal("http"))

		var updatedRoute gatewayapi.HTTPRoute
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updatedRoute)).To(Succeed())
		Expect(updatedRoute.Status.Parents).To(HaveLen(1))
		condition := kube_apimeta.FindStatusCondition(updatedRoute.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionResolvedRefs))
		Expect(condition).ToNot(BeNil())
		Expect(condition.Status).To(Equal(kube_meta.ConditionTrue))
	})

	It("applies a parentRef sectionName as the port name", func() {
		ms := &meshservice_k8s.MeshService{
			ObjectMeta: kube_meta.ObjectMeta{Name: "backend", Namespace: "kuma-demo"},
			Spec: &meshservice_api.MeshService{
				Ports: []meshservice_api.Port{
					{Port: 80, Name: pointer.To("http")},
					{Port: 443, Name: pointer.To("https")},
				},
			},
		}
		route := newRoute(
			withSectionName(meshServiceParentRef("backend"), "https"),
		)

		client := newClientBuilder(ms, route)
		reconciler.Client = client

		_, err := reconciler.Reconcile(context.Background(), kube_ctrl.Request{
			NamespacedName: kube_client.ObjectKeyFromObject(route),
		})
		Expect(err).ToNot(HaveOccurred())

		routes := &meshhttproute_k8s.MeshHTTPRouteList{}
		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(HaveLen(1))

		spec := routes.Items[0].Spec
		Expect(spec).ToNot(BeNil())
		Expect(*spec.To).To(HaveLen(1))
		Expect(*(*spec.To)[0].TargetRef.SectionName).To(Equal("https"))
	})

	It("reports Accepted=False when the parentRef names a port the MeshService does not have", func() {
		ms := &meshservice_k8s.MeshService{
			ObjectMeta: kube_meta.ObjectMeta{Name: "backend", Namespace: "kuma-demo"},
			Spec: &meshservice_api.MeshService{
				Ports: []meshservice_api.Port{{Port: 80, Name: pointer.To("http")}},
			},
		}
		route := newRoute(
			withSectionName(meshServiceParentRef("backend"), "grpc"),
		)

		client := newClientBuilder(ms, route)
		reconciler.Client = client

		_, err := reconciler.Reconcile(context.Background(), kube_ctrl.Request{
			NamespacedName: kube_client.ObjectKeyFromObject(route),
		})
		Expect(err).ToNot(HaveOccurred())

		routes := &meshhttproute_k8s.MeshHTTPRouteList{}
		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(BeEmpty())

		var updatedRoute gatewayapi.HTTPRoute
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updatedRoute)).To(Succeed())
		Expect(updatedRoute.Status.Parents).To(HaveLen(1))
		accepted := kube_apimeta.FindStatusCondition(updatedRoute.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionAccepted))
		Expect(accepted).ToNot(BeNil())
		Expect(accepted.Status).To(Equal(kube_meta.ConditionFalse))
		Expect(accepted.Reason).To(Equal(string(gatewayapi_v1.RouteReasonNoMatchingParent)))
	})

	It("stays quiet when the MeshService has no ports and the parentRef names none", func() {
		ms := &meshservice_k8s.MeshService{
			ObjectMeta: kube_meta.ObjectMeta{Name: "backend", Namespace: "kuma-demo"},
			Spec:       &meshservice_api.MeshService{},
		}
		route := newRoute(meshServiceParentRef("backend"))

		client := newClientBuilder(ms, route)
		reconciler.Client = client

		_, err := reconciler.Reconcile(context.Background(), kube_ctrl.Request{
			NamespacedName: kube_client.ObjectKeyFromObject(route),
		})
		Expect(err).ToNot(HaveOccurred())

		var updatedRoute gatewayapi.HTTPRoute
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updatedRoute)).To(Succeed())
		Expect(updatedRoute.Status.Parents).To(HaveLen(1))
		accepted := kube_apimeta.FindStatusCondition(updatedRoute.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionAccepted))
		Expect(accepted).ToNot(BeNil())
		Expect(accepted.Status).To(Equal(kube_meta.ConditionTrue))
	})

	It("reports Accepted=False and generates no MeshHTTPRoute when the MeshService does not exist", func() {
		route := newRoute(meshServiceParentRef("missing"))

		client := newClientBuilder(route)
		reconciler.Client = client

		_, err := reconciler.Reconcile(context.Background(), kube_ctrl.Request{
			NamespacedName: kube_client.ObjectKeyFromObject(route),
		})
		Expect(err).ToNot(HaveOccurred())

		routes := &meshhttproute_k8s.MeshHTTPRouteList{}
		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(BeEmpty())

		var updatedRoute gatewayapi.HTTPRoute
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updatedRoute)).To(Succeed())
		Expect(updatedRoute.Status.Parents).To(HaveLen(1))
		accepted := kube_apimeta.FindStatusCondition(updatedRoute.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionAccepted))
		Expect(accepted).ToNot(BeNil())
		Expect(accepted.Status).To(Equal(kube_meta.ConditionFalse))
		Expect(accepted.Reason).To(Equal(string(gatewayapi_v1.RouteReasonNoMatchingParent)))

		// backendRefs resolve fine; a missing parent is not a backend problem
		condition := kube_apimeta.FindStatusCondition(updatedRoute.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionResolvedRefs))
		Expect(condition).ToNot(BeNil())
		Expect(condition.Status).To(Equal(kube_meta.ConditionTrue))
	})
})

var _ = Describe("meshServicesOfRoute", func() {
	meshServiceRef := func(namespace, name string) *gatewayapi.BackendObjectReference {
		group := gatewayapi.Group(meshservice_k8s.GroupVersion.Group)
		kind := gatewayapi.Kind("MeshService")
		ref := &gatewayapi.BackendObjectReference{
			Group: &group,
			Kind:  &kind,
			Name:  gatewayapi.ObjectName(name),
		}
		if namespace != "" {
			ns := gatewayapi.Namespace(namespace)
			ref.Namespace = &ns
		}
		return ref
	}

	parentRef := func(ref *gatewayapi.BackendObjectReference) gatewayapi.ParentReference {
		return gatewayapi.ParentReference{
			Group:     ref.Group,
			Kind:      ref.Kind,
			Namespace: ref.Namespace,
			Name:      ref.Name,
		}
	}

	It("indexes parentRefs, backendRefs and request-mirror backendRefs, defaulting namespace to the route", func() {
		route := &gatewayapi.HTTPRoute{
			ObjectMeta: kube_meta.ObjectMeta{Name: "my-route", Namespace: "kuma-demo"},
			Spec: gatewayapi.HTTPRouteSpec{
				CommonRouteSpec: gatewayapi.CommonRouteSpec{
					ParentRefs: []gatewayapi.ParentReference{
						parentRef(meshServiceRef("", "parent-ms")),
					},
				},
				Rules: []gatewayapi.HTTPRouteRule{
					{
						BackendRefs: []gatewayapi.HTTPBackendRef{
							{BackendRef: gatewayapi.BackendRef{BackendObjectReference: *meshServiceRef("other-ns", "backend-ms")}},
						},
						Filters: []gatewayapi.HTTPRouteFilter{
							{
								Type: gatewayapi_v1.HTTPRouteFilterRequestMirror,
								RequestMirror: &gatewayapi.HTTPRequestMirrorFilter{
									BackendRef: *meshServiceRef("mirror-ns", "mirror-ms"),
								},
							},
						},
					},
				},
			},
		}

		names := meshServicesOfRoute(route)
		Expect(names).To(ConsistOf(
			"kuma-demo/parent-ms",
			"other-ns/backend-ms",
			"mirror-ns/mirror-ms",
		))
	})
})
