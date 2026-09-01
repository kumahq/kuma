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
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi_beta "sigs.k8s.io/gateway-api/apis/v1beta1"

	bootstrap_k8s "github.com/kumahq/kuma/v3/pkg/plugins/bootstrap/k8s"
	meshhttproute_k8s "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/k8s/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var _ = Describe("GRPCRouteReconciler", func() {
	newServiceParentRef := func(namespace, name string) gatewayapi.ParentReference {
		group := gatewayapi.Group("")
		kind := gatewayapi.Kind("Service")
		ns := gatewayapi.Namespace(namespace)
		return gatewayapi.ParentReference{
			Group:     &group,
			Kind:      &kind,
			Namespace: &ns,
			Name:      gatewayapi.ObjectName(name),
		}
	}

	newBackendRef := func(name string) gatewayapi.GRPCBackendRef {
		group := gatewayapi.Group("")
		kind := gatewayapi.Kind("Service")
		return gatewayapi.GRPCBackendRef{
			BackendRef: gatewayapi.BackendRef{
				BackendObjectReference: gatewayapi.BackendObjectReference{
					Group: &group,
					Kind:  &kind,
					Name:  gatewayapi.ObjectName(name),
					Port:  pointer.To(gatewayapi.PortNumber(80)),
				},
			},
		}
	}

	newGRPCRoute := func(namespace, name string, parentRefs ...gatewayapi.ParentReference) *gatewayapi.GRPCRoute {
		return &gatewayapi.GRPCRoute{
			ObjectMeta: kube_meta.ObjectMeta{Name: name, Namespace: namespace},
			Spec: gatewayapi.GRPCRouteSpec{
				CommonRouteSpec: gatewayapi.CommonRouteSpec{ParentRefs: parentRefs},
				Rules: []gatewayapi.GRPCRouteRule{{
					BackendRefs: []gatewayapi.GRPCBackendRef{newBackendRef("backend")},
				}},
			},
		}
	}

	var namespace *kube_core.Namespace
	var service *kube_core.Service
	var grpcReconciler *GRPCRouteReconciler
	var httpReconciler *HTTPRouteReconciler
	var newClient func(...kube_client.Object) kube_client.Client

	BeforeEach(func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		namespace = &kube_core.Namespace{ObjectMeta: kube_meta.ObjectMeta{Name: "kuma-demo"}}
		service = &kube_core.Service{
			ObjectMeta: kube_meta.ObjectMeta{Name: "backend", Namespace: "kuma-demo"},
			Spec: kube_core.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports:     []kube_core.ServicePort{{Name: "grpc", Port: 80}},
			},
		}

		newClient = func(objs ...kube_client.Object) kube_client.Client {
			return kube_client_fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&gatewayapi.GRPCRoute{}, &gatewayapi_beta.HTTPRoute{}).
				WithIndex(&gatewayapi.GRPCRoute{}, servicesOfRouteField, servicesOfGRPCRoute).
				WithIndex(&gatewayapi.GRPCRoute{}, meshServicesOfRouteField, meshServicesOfGRPCRoute).
				WithObjects(append([]kube_client.Object{namespace}, objs...)...).
				Build()
		}

		grpcReconciler = &GRPCRouteReconciler{
			Log:             logr.Discard(),
			Scheme:          scheme,
			TypeRegistry:    k8s_registry.Global(),
			SystemNamespace: "kuma-system",
		}
		httpReconciler = &HTTPRouteReconciler{
			Log:             logr.Discard(),
			Scheme:          scheme,
			TypeRegistry:    k8s_registry.Global(),
			SystemNamespace: "kuma-system",
		}
	})

	It("generates a MeshHTTPRoute and reports Accepted and ResolvedRefs on its parent", func() {
		route := newGRPCRoute("kuma-demo", "my-route", newServiceParentRef("kuma-demo", "backend"))
		client := newClient(service, route)
		grpcReconciler.Client = client

		_, err := grpcReconciler.Reconcile(context.Background(), kube_ctrl.Request{NamespacedName: kube_client.ObjectKeyFromObject(route)})
		Expect(err).ToNot(HaveOccurred())

		routes := &meshhttproute_k8s.MeshHTTPRouteList{}
		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(HaveLen(1))
		Expect(routes.Items[0].Name).To(Equal(generatedMeshHTTPRouteName(sourceRouteKindGRPCRoute, route.Namespace, route.Name, "Service", service.Namespace, service.Name)))

		var updated gatewayapi.GRPCRoute
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updated)).To(Succeed())
		Expect(updated.Status.Parents).To(HaveLen(1))
		accepted := kube_apimeta.FindStatusCondition(updated.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionAccepted))
		resolved := kube_apimeta.FindStatusCondition(updated.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionResolvedRefs))
		Expect(accepted).ToNot(BeNil())
		Expect(accepted.Status).To(Equal(kube_meta.ConditionTrue))
		Expect(resolved).ToNot(BeNil())
		Expect(resolved.Status).To(Equal(kube_meta.ConditionTrue))
	})

	It("reports a missing parent service and unresolved backend refs in status", func() {
		route := newGRPCRoute("kuma-demo", "missing-parent", newServiceParentRef("kuma-demo", "missing-parent"))
		client := newClient(route)
		grpcReconciler.Client = client

		_, err := grpcReconciler.Reconcile(context.Background(), kube_ctrl.Request{NamespacedName: kube_client.ObjectKeyFromObject(route)})
		Expect(err).ToNot(HaveOccurred())

		var updated gatewayapi.GRPCRoute
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updated)).To(Succeed())
		Expect(updated.Status.Parents).To(HaveLen(1))
		accepted := kube_apimeta.FindStatusCondition(updated.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionAccepted))
		resolved := kube_apimeta.FindStatusCondition(updated.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionResolvedRefs))
		Expect(accepted).ToNot(BeNil())
		Expect(accepted.Status).To(Equal(kube_meta.ConditionFalse))
		Expect(accepted.Reason).To(Equal(string(gatewayapi.RouteReasonNoMatchingParent)))
		Expect(resolved).ToNot(BeNil())
		Expect(resolved.Status).To(Equal(kube_meta.ConditionFalse))
		Expect(resolved.Reason).To(Equal(string(gatewayapi.RouteReasonBackendNotFound)))
	})

	It("reports unsupported extension filters and removes generated routes", func() {
		route := newGRPCRoute("kuma-demo", "extension-filter", newServiceParentRef("kuma-demo", "backend"))
		extensionGroup := gatewayapi.Group("example.com")
		extensionKind := gatewayapi.Kind("AuthFilter")
		route.Spec.Rules[0].Filters = []gatewayapi.GRPCRouteFilter{{
			Type: gatewayapi.GRPCRouteFilterExtensionRef,
			ExtensionRef: &gatewayapi.LocalObjectReference{
				Group: extensionGroup,
				Kind:  extensionKind,
				Name:  "authn",
			},
		}}
		existingRoute := &meshhttproute_k8s.MeshHTTPRoute{
			ObjectMeta: kube_meta.ObjectMeta{
				Name:      generatedMeshHTTPRouteName(sourceRouteKindGRPCRoute, route.Namespace, route.Name, "Service", service.Namespace, service.Name),
				Namespace: route.Namespace,
				Labels: map[string]string{
					common.OwnerLabel: common.OwnerLabelValue(sourceRouteKindGRPCRoute, kube_client.ObjectKeyFromObject(route)),
				},
			},
		}
		client := newClient(service, route, existingRoute)
		grpcReconciler.Client = client

		_, err := grpcReconciler.Reconcile(context.Background(), kube_ctrl.Request{NamespacedName: kube_client.ObjectKeyFromObject(route)})
		Expect(err).ToNot(HaveOccurred())

		routes := &meshhttproute_k8s.MeshHTTPRouteList{}
		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(BeEmpty())

		var updated gatewayapi.GRPCRoute
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updated)).To(Succeed())
		Expect(updated.Status.Parents).To(HaveLen(1))
		accepted := kube_apimeta.FindStatusCondition(updated.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionAccepted))
		Expect(accepted).ToNot(BeNil())
		Expect(accepted.Status).To(Equal(kube_meta.ConditionFalse))
		Expect(accepted.Reason).To(Equal(string(gatewayapi.RouteReasonUnsupportedValue)))
	})

	It("does not collide with an HTTPRoute that shares namespace name and parent", func() {
		parentRef := newServiceParentRef("kuma-demo", "backend")
		httpRoute := &gatewayapi_beta.HTTPRoute{
			ObjectMeta: kube_meta.ObjectMeta{Name: "shared", Namespace: "kuma-demo"},
			Spec: gatewayapi_beta.HTTPRouteSpec{
				CommonRouteSpec: gatewayapi_beta.CommonRouteSpec{ParentRefs: []gatewayapi_beta.ParentReference{gatewayapi_beta.ParentReference(parentRef)}},
				Rules: []gatewayapi_beta.HTTPRouteRule{{
					Matches: []gatewayapi_beta.HTTPRouteMatch{{
						Path: &gatewayapi_beta.HTTPPathMatch{
							Type:  pointer.To(gatewayapi_beta.PathMatchType("PathPrefix")),
							Value: pointer.To("/"),
						},
					}},
					BackendRefs: []gatewayapi_beta.HTTPBackendRef{{
						BackendRef: gatewayapi_beta.BackendRef{
							BackendObjectReference: gatewayapi_beta.BackendObjectReference{
								Name: "backend",
								Port: pointer.To(gatewayapi_beta.PortNumber(80)),
							},
							Weight: pointer.To(int32(1)),
						},
					}},
				}},
			},
		}
		grpcRoute := newGRPCRoute("kuma-demo", "shared", parentRef)

		client := newClient(service, httpRoute, grpcRoute)
		httpReconciler.Client = client
		grpcReconciler.Client = client

		_, err := httpReconciler.Reconcile(context.Background(), kube_ctrl.Request{NamespacedName: kube_client.ObjectKeyFromObject(httpRoute)})
		Expect(err).ToNot(HaveOccurred())
		_, err = grpcReconciler.Reconcile(context.Background(), kube_ctrl.Request{NamespacedName: kube_client.ObjectKeyFromObject(grpcRoute)})
		Expect(err).ToNot(HaveOccurred())
		_, err = httpReconciler.Reconcile(context.Background(), kube_ctrl.Request{NamespacedName: kube_client.ObjectKeyFromObject(httpRoute)})
		Expect(err).ToNot(HaveOccurred())

		routes := &meshhttproute_k8s.MeshHTTPRouteList{}
		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(HaveLen(2))
		Expect(routes.Items[0].Name).ToNot(Equal(routes.Items[1].Name))
	})
})
