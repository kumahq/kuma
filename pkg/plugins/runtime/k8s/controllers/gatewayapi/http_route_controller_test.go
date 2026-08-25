package gatewayapi

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_event "sigs.k8s.io/controller-runtime/pkg/event"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
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
			Name: "my-route", Namespace: routeNamespace,
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

		namespace = &kube_core.Namespace{Name: "kuma-demo"}

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
			Name: "backend", Namespace: "kuma-demo",
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
			Name: "backend", Namespace: "kuma-demo",
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
			Name: "backend", Namespace: "kuma-demo",
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
			Name: "backend", Namespace: "kuma-demo",
			Spec: &meshservice_api.MeshService{},
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

var _ = Describe("HTTPRouteReconciler.Reconcile with a Service parentRef", func() {
	const routeNamespace = "kuma-demo"

	serviceParentRef := func() gatewayapi.ParentReference {
		group := gatewayapi.Group("")
		kind := gatewayapi.Kind("Service")
		ns := gatewayapi.Namespace(routeNamespace)
		return gatewayapi.ParentReference{
			Group:     &group,
			Kind:      &kind,
			Namespace: &ns,
			Name:      gatewayapi.ObjectName("backend"),
		}
	}

	withSectionName := func(ref gatewayapi.ParentReference, sectionName string) gatewayapi.ParentReference {
		ref.SectionName = pointer.To(gatewayapi.SectionName(sectionName))
		return ref
	}

	newRoute := func(parentRefs ...gatewayapi.ParentReference) *gatewayapi.HTTPRoute {
		return &gatewayapi.HTTPRoute{
			Name: "my-route", Namespace: routeNamespace,
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

		namespace = &kube_core.Namespace{Name: routeNamespace}

		newClientBuilder = func(objs ...kube_client.Object) kube_client.Client {
			return kube_client_fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&gatewayapi.HTTPRoute{}).
				WithIndex(&gatewayapi.HTTPRoute{}, servicesOfRouteField, servicesOfRoute).
				WithObjects(append([]kube_client.Object{namespace}, objs...)...).
				Build()
		}

		reconciler = &HTTPRouteReconciler{
			Log:             logr.Discard(),
			TypeRegistry:    k8s_registry.Global(),
			SystemNamespace: "kuma-system",
		}
	})

	It("applies a parentRef sectionName as the Service port name", func() {
		svc := &kube_core.Service{
			Name: "backend", Namespace: routeNamespace,
			Spec: kube_core.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []kube_core.ServicePort{
					{Name: "http", Port: 80},
					{Name: "https", Port: 443},
				},
			},
		}
		route := newRoute(withSectionName(serviceParentRef(), "https"))

		client := newClientBuilder(svc, route)
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

		var updatedRoute gatewayapi.HTTPRoute
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updatedRoute)).To(Succeed())
		Expect(updatedRoute.Status.Parents).To(HaveLen(1))
		accepted := kube_apimeta.FindStatusCondition(updatedRoute.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionAccepted))
		Expect(accepted).ToNot(BeNil())
		Expect(accepted.Status).To(Equal(kube_meta.ConditionTrue))
	})

	It("merges multiple section-specific parentRefs for the same Service", func() {
		svc := &kube_core.Service{
			Name: "backend", Namespace: routeNamespace,
			Spec: kube_core.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []kube_core.ServicePort{
					{Name: "http", Port: 80},
					{Name: "https", Port: 443},
				},
			},
		}
		route := newRoute(
			withSectionName(serviceParentRef(), "http"),
			withSectionName(serviceParentRef(), "https"),
		)

		client := newClientBuilder(svc, route)
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
		Expect(*spec.To).To(HaveLen(2))
		Expect(*(*spec.To)[0].TargetRef.SectionName).To(Equal("http"))
		Expect(*(*spec.To)[1].TargetRef.SectionName).To(Equal("https"))

		var updatedRoute gatewayapi.HTTPRoute
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updatedRoute)).To(Succeed())
		Expect(updatedRoute.Status.Parents).To(HaveLen(2))
		for _, parent := range updatedRoute.Status.Parents {
			accepted := kube_apimeta.FindStatusCondition(parent.Conditions, string(gatewayapi.RouteConditionAccepted))
			Expect(accepted).ToNot(BeNil())
			Expect(accepted.Status).To(Equal(kube_meta.ConditionTrue))
		}
	})

	It("reports Accepted=False when the parentRef names a Service port the Service does not have", func() {
		svc := &kube_core.Service{
			Name: "backend", Namespace: routeNamespace,
			Spec: kube_core.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports:     []kube_core.ServicePort{{Name: "http", Port: 80}},
			},
		}
		route := newRoute(withSectionName(serviceParentRef(), "grpc"))

		client := newClientBuilder(svc, route)
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

	It("reports Accepted=False after deleting a Service referenced only as a parentRef", func() {
		svc := &kube_core.Service{
			Name: "backend", Namespace: routeNamespace,
			Spec: kube_core.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports:     []kube_core.ServicePort{{Name: "http", Port: 80}},
			},
		}
		route := newRoute(withSectionName(serviceParentRef(), "http"))

		client := newClientBuilder(svc, route)
		reconciler.Client = client

		_, err := reconciler.Reconcile(context.Background(), kube_ctrl.Request{
			NamespacedName: kube_client.ObjectKeyFromObject(route),
		})
		Expect(err).ToNot(HaveOccurred())

		routes := &meshhttproute_k8s.MeshHTTPRouteList{}
		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(HaveLen(1))

		deletedSvc := svc.DeepCopy()
		Expect(client.Delete(context.Background(), svc)).To(Succeed())

		requests := routesForService(logr.Discard(), client)(context.Background(), deletedSvc)
		Expect(requests).To(ConsistOf(kube_ctrl.Request{
			NamespacedName: kube_client.ObjectKeyFromObject(route),
		}))

		_, err = reconciler.Reconcile(context.Background(), requests[0])
		Expect(err).ToNot(HaveOccurred())

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

	It("requeues a parent-only route when a Service update makes its sectionName valid", func() {
		svc := &kube_core.Service{
			Name: "backend", Namespace: routeNamespace,
			Spec: kube_core.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports:     []kube_core.ServicePort{{Name: "http", Port: 80}},
			},
		}
		route := newRoute(withSectionName(serviceParentRef(), "grpc"))

		client := newClientBuilder(svc, route)
		reconciler.Client = client

		_, err := reconciler.Reconcile(context.Background(), kube_ctrl.Request{
			NamespacedName: kube_client.ObjectKeyFromObject(route),
		})
		Expect(err).ToNot(HaveOccurred())

		routes := &meshhttproute_k8s.MeshHTTPRouteList{}
		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(BeEmpty())

		var updatedSvc kube_core.Service
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(svc), &updatedSvc)).To(Succeed())
		updatedSvc.Spec.Ports = []kube_core.ServicePort{{Name: "grpc", Port: 50051}}
		Expect(client.Update(context.Background(), &updatedSvc)).To(Succeed())

		requests := routesForService(logr.Discard(), client)(context.Background(), &updatedSvc)
		Expect(requests).To(ConsistOf(kube_ctrl.Request{
			NamespacedName: kube_client.ObjectKeyFromObject(route),
		}))

		_, err = reconciler.Reconcile(context.Background(), requests[0])
		Expect(err).ToNot(HaveOccurred())

		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(HaveLen(1))
		spec := routes.Items[0].Spec
		Expect(spec).ToNot(BeNil())
		Expect(*spec.To).To(HaveLen(1))
		Expect(*(*spec.To)[0].TargetRef.SectionName).To(Equal("grpc"))

		var updatedRoute gatewayapi.HTTPRoute
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updatedRoute)).To(Succeed())
		Expect(updatedRoute.Status.Parents).To(HaveLen(1))
		accepted := kube_apimeta.FindStatusCondition(updatedRoute.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionAccepted))
		Expect(accepted).ToNot(BeNil())
		Expect(accepted.Status).To(Equal(kube_meta.ConditionTrue))
	})

	It("reports Accepted=False and generates no MeshHTTPRoute for a headless Service parentRef", func() {
		svc := &kube_core.Service{
			Name: "backend", Namespace: routeNamespace,
			Spec: kube_core.ServiceSpec{
				ClusterIP: kube_core.ClusterIPNone,
				Ports:     []kube_core.ServicePort{{Name: "http", Port: 80}},
			},
		}
		route := newRoute(withSectionName(serviceParentRef(), "http"))

		client := newClientBuilder(svc, route)
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
		Expect(accepted.Message).To(ContainSubstring("has no MeshService to attach to"))
	})

	It("requeues a parent-only route when a headless Service is replaced by a ClusterIP Service", func() {
		svc := &kube_core.Service{
			Name: "backend", Namespace: routeNamespace,
			Spec: kube_core.ServiceSpec{
				ClusterIP: kube_core.ClusterIPNone,
				Ports:     []kube_core.ServicePort{{Name: "http", Port: 80}},
			},
		}
		route := newRoute(withSectionName(serviceParentRef(), "http"))

		client := newClientBuilder(svc, route)
		reconciler.Client = client

		req := kube_ctrl.Request{NamespacedName: kube_client.ObjectKeyFromObject(route)}
		_, err := reconciler.Reconcile(context.Background(), req)
		Expect(err).ToNot(HaveOccurred())

		routes := &meshhttproute_k8s.MeshHTTPRouteList{}
		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(BeEmpty())

		recreatedSvc := svc.DeepCopy()
		recreatedSvc.Spec.ClusterIP = "10.0.0.9"
		Expect(client.Delete(context.Background(), svc)).To(Succeed())
		recreatedSvc.ResourceVersion = ""
		Expect(client.Create(context.Background(), recreatedSvc)).To(Succeed())

		requests := routesForService(logr.Discard(), client)(context.Background(), recreatedSvc)
		Expect(requests).To(ConsistOf(req))

		_, err = reconciler.Reconcile(context.Background(), req)
		Expect(err).ToNot(HaveOccurred())

		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(HaveLen(1))
		spec := routes.Items[0].Spec
		Expect(spec).ToNot(BeNil())
		Expect(*spec.To).To(HaveLen(1))
		Expect(*(*spec.To)[0].TargetRef.SectionName).To(Equal("http"))

		var updatedRoute gatewayapi.HTTPRoute
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updatedRoute)).To(Succeed())
		Expect(updatedRoute.Status.Parents).To(HaveLen(1))
		accepted := kube_apimeta.FindStatusCondition(updatedRoute.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionAccepted))
		Expect(accepted).ToNot(BeNil())
		Expect(accepted.Status).To(Equal(kube_meta.ConditionTrue))
	})
})

var _ = Describe("HTTPRouteReconciler.Reconcile with cross-namespace backendRefs", func() {
	serviceBackendRef := func(namespace, name string, port gatewayapi.PortNumber) gatewayapi.HTTPBackendRef {
		group := gatewayapi.Group("")
		kind := gatewayapi.Kind("Service")
		ns := gatewayapi.Namespace(namespace)
		weight := int32(1)
		return gatewayapi.HTTPBackendRef{
			Group:     &group,
			Kind:      &kind,
			Namespace: &ns,
			Name:      gatewayapi.ObjectName(name),
			Port:      &port,
			Weight:    &weight,
		}
	}

	referenceGrant := func(grantNamespace, routeNamespace, targetName string) *gatewayapi.ReferenceGrant {
		return &gatewayapi.ReferenceGrant{
			Name: "allow-route", Namespace: grantNamespace,
			Spec: gatewayapi.ReferenceGrantSpec{
				From: []gatewayapi.ReferenceGrantFrom{{
					Group:     gatewayapi.Group(gatewayapi.GroupVersion.Group),
					Kind:      gatewayapi.Kind("HTTPRoute"),
					Namespace: gatewayapi.Namespace(routeNamespace),
				}},
				To: []gatewayapi.ReferenceGrantTo{{
					Group: gatewayapi.Group(""),
					Kind:  gatewayapi.Kind("Service"),
					Name:  pointer.To(gatewayapi.ObjectName(targetName)),
				}},
			},
		}
	}

	parentRef := func(namespace, name string) gatewayapi.ParentReference {
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

	newRoute := func() *gatewayapi.HTTPRoute {
		return &gatewayapi.HTTPRoute{
			Name: "my-route", Namespace: "route-ns",
			Spec: gatewayapi.HTTPRouteSpec{
				CommonRouteSpec: gatewayapi.CommonRouteSpec{
					ParentRefs: []gatewayapi.ParentReference{parentRef("route-ns", "frontend")},
				},
				Rules: []gatewayapi.HTTPRouteRule{{
					BackendRefs: []gatewayapi.HTTPBackendRef{
						serviceBackendRef("backend-ns", "backend", 8080),
					},
				}},
			},
		}
	}

	var reconciler *HTTPRouteReconciler
	var newClientBuilder func(objs ...kube_client.Object) kube_client.Client

	BeforeEach(func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		newClientBuilder = func(objs ...kube_client.Object) kube_client.Client {
			base := []kube_client.Object{
				&kube_core.Namespace{Name: "route-ns"},
				&kube_core.Namespace{Name: "backend-ns"},
			}
			return kube_client_fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&gatewayapi.HTTPRoute{}).
				WithIndex(&gatewayapi.HTTPRoute{}, servicesOfRouteField, servicesOfRoute).
				WithIndex(&gatewayapi.HTTPRoute{}, meshServicesOfRouteField, meshServicesOfRoute).
				WithObjects(append(base, objs...)...).
				Build()
		}

		reconciler = &HTTPRouteReconciler{
			Log:             logr.Discard(),
			TypeRegistry:    k8s_registry.Global(),
			SystemNamespace: "kuma-system",
		}
	})

	It("reports RefNotPermitted and omits the denied backend targetRef when no ReferenceGrant exists", func() {
		frontend := &kube_core.Service{
			Name: "frontend", Namespace: "route-ns",
			Spec: kube_core.ServiceSpec{
				Ports: []kube_core.ServicePort{{Name: "http", Port: 80}},
			},
		}
		backend := &kube_core.Service{
			Name: "backend", Namespace: "backend-ns",
			Spec: kube_core.ServiceSpec{
				Ports: []kube_core.ServicePort{{Name: "http", Port: 8080}},
			},
		}
		route := newRoute()

		client := newClientBuilder(frontend, backend, route)
		reconciler.Client = client

		_, err := reconciler.Reconcile(context.Background(), kube_ctrl.Request{
			NamespacedName: kube_client.ObjectKeyFromObject(route),
		})
		Expect(err).ToNot(HaveOccurred())

		routes := &meshhttproute_k8s.MeshHTTPRouteList{}
		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(HaveLen(1))
		backendRefs := pointer.Deref((*routes.Items[0].Spec.To)[0].Rules[0].Default.BackendRefs)
		Expect(backendRefs).To(BeEmpty())

		var updatedRoute gatewayapi.HTTPRoute
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updatedRoute)).To(Succeed())
		resolvedRefs := kube_apimeta.FindStatusCondition(updatedRoute.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionResolvedRefs))
		Expect(resolvedRefs).ToNot(BeNil())
		Expect(resolvedRefs.Status).To(Equal(kube_meta.ConditionFalse))
		Expect(resolvedRefs.Reason).To(Equal(string(gatewayapi.RouteReasonRefNotPermitted)))
		Expect(resolvedRefs.Message).To(ContainSubstring("backend-ns/backend"))
	})

	It("requeues and allows the route after grant creation, then denies it again after grant deletion", func() {
		frontend := &kube_core.Service{
			Name: "frontend", Namespace: "route-ns",
			Spec: kube_core.ServiceSpec{
				Ports: []kube_core.ServicePort{{Name: "http", Port: 80}},
			},
		}
		backend := &kube_core.Service{
			Name: "backend", Namespace: "backend-ns",
			Spec: kube_core.ServiceSpec{
				Ports: []kube_core.ServicePort{{Name: "http", Port: 8080}},
			},
		}
		route := newRoute()

		client := newClientBuilder(frontend, backend, route)
		reconciler.Client = client

		req := kube_ctrl.Request{NamespacedName: kube_client.ObjectKeyFromObject(route)}
		_, err := reconciler.Reconcile(context.Background(), req)
		Expect(err).ToNot(HaveOccurred())

		grant := referenceGrant("backend-ns", "route-ns", "backend")
		Expect(client.Create(context.Background(), grant)).To(Succeed())

		requests := routesForReferenceGrant(logr.Discard(), client)(context.Background(), grant)
		Expect(requests).To(ConsistOf(req))

		_, err = reconciler.Reconcile(context.Background(), req)
		Expect(err).ToNot(HaveOccurred())

		routes := &meshhttproute_k8s.MeshHTTPRouteList{}
		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(HaveLen(1))
		backendRefs := pointer.Deref((*routes.Items[0].Spec.To)[0].Rules[0].Default.BackendRefs)
		Expect(backendRefs).To(HaveLen(1))
		Expect(backendRefs[0].TargetRef.SectionName).To(Equal(pointer.To("http")))

		var updatedRoute gatewayapi.HTTPRoute
		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updatedRoute)).To(Succeed())
		resolvedRefs := kube_apimeta.FindStatusCondition(updatedRoute.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionResolvedRefs))
		Expect(resolvedRefs).ToNot(BeNil())
		Expect(resolvedRefs.Status).To(Equal(kube_meta.ConditionTrue))

		deletedGrant := grant.DeepCopy()
		Expect(client.Delete(context.Background(), grant)).To(Succeed())
		requests = routesForReferenceGrant(logr.Discard(), client)(context.Background(), deletedGrant)
		Expect(requests).To(ConsistOf(req))

		_, err = reconciler.Reconcile(context.Background(), req)
		Expect(err).ToNot(HaveOccurred())

		routes = &meshhttproute_k8s.MeshHTTPRouteList{}
		Expect(client.List(context.Background(), routes)).To(Succeed())
		Expect(routes.Items).To(HaveLen(1))
		backendRefs = pointer.Deref((*routes.Items[0].Spec.To)[0].Rules[0].Default.BackendRefs)
		Expect(backendRefs).To(BeEmpty())

		Expect(client.Get(context.Background(), kube_client.ObjectKeyFromObject(route), &updatedRoute)).To(Succeed())
		resolvedRefs = kube_apimeta.FindStatusCondition(updatedRoute.Status.Parents[0].Conditions, string(gatewayapi.RouteConditionResolvedRefs))
		Expect(resolvedRefs).ToNot(BeNil())
		Expect(resolvedRefs.Status).To(Equal(kube_meta.ConditionFalse))
		Expect(resolvedRefs.Reason).To(Equal(string(gatewayapi.RouteReasonRefNotPermitted)))
	})
})

var _ = Describe("routesForReferenceGrant", func() {
	serviceBackendRef := func(name string) gatewayapi.HTTPBackendRef {
		group := gatewayapi.Group("")
		kind := gatewayapi.Kind("Service")
		ns := gatewayapi.Namespace("backend-ns")
		port := gatewayapi.PortNumber(80)
		weight := int32(1)
		return gatewayapi.HTTPBackendRef{
			Group:     &group,
			Kind:      &kind,
			Namespace: &ns,
			Name:      gatewayapi.ObjectName(name),
			Port:      &port,
			Weight:    &weight,
		}
	}

	It("requeues only routes matched by the grant source and destination rules", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		matchingRoute := &gatewayapi.HTTPRoute{
			Name: "matching", Namespace: "route-ns",
			Spec: gatewayapi.HTTPRouteSpec{
				Rules: []gatewayapi.HTTPRouteRule{{
					BackendRefs: []gatewayapi.HTTPBackendRef{serviceBackendRef("backend")},
				}},
			},
		}
		sameNamespaceRoute := &gatewayapi.HTTPRoute{
			Name: "same-ns", Namespace: "backend-ns",
			Spec: gatewayapi.HTTPRouteSpec{
				Rules: []gatewayapi.HTTPRouteRule{{
					BackendRefs: []gatewayapi.HTTPBackendRef{serviceBackendRef("backend")},
				}},
			},
		}
		wrongBackendRoute := &gatewayapi.HTTPRoute{
			Name: "wrong-backend", Namespace: "route-ns",
			Spec: gatewayapi.HTTPRouteSpec{
				Rules: []gatewayapi.HTTPRouteRule{{
					BackendRefs: []gatewayapi.HTTPBackendRef{serviceBackendRef("other-backend")},
				}},
			},
		}
		grant := &gatewayapi.ReferenceGrant{
			Name: "allow-route", Namespace: "backend-ns",
			Spec: gatewayapi.ReferenceGrantSpec{
				From: []gatewayapi.ReferenceGrantFrom{{
					Group:     gatewayapi.Group(gatewayapi.GroupVersion.Group),
					Kind:      gatewayapi.Kind("HTTPRoute"),
					Namespace: gatewayapi.Namespace("route-ns"),
				}},
				To: []gatewayapi.ReferenceGrantTo{{
					Group: gatewayapi.Group(""),
					Kind:  gatewayapi.Kind("Service"),
					Name:  pointer.To(gatewayapi.ObjectName("backend")),
				}},
			},
		}

		client := kube_client_fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(matchingRoute, sameNamespaceRoute, wrongBackendRoute).
			Build()

		requests := routesForReferenceGrant(logr.Discard(), client)(context.Background(), grant)
		Expect(requests).To(ConsistOf(kube_ctrl.Request{
			NamespacedName: kube_client.ObjectKeyFromObject(matchingRoute),
		}))
	})

	It("requeues routes matched by the previous grant spec when an update revokes access", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		matchingRoute := &gatewayapi.HTTPRoute{
			Name: "matching", Namespace: "route-ns",
			Spec: gatewayapi.HTTPRouteSpec{
				Rules: []gatewayapi.HTTPRouteRule{{
					BackendRefs: []gatewayapi.HTTPBackendRef{serviceBackendRef("backend")},
				}},
			},
		}
		oldGrant := &gatewayapi.ReferenceGrant{
			Name: "allow-route", Namespace: "backend-ns",
			Spec: gatewayapi.ReferenceGrantSpec{
				From: []gatewayapi.ReferenceGrantFrom{{
					Group:     gatewayapi.Group(gatewayapi.GroupVersion.Group),
					Kind:      gatewayapi.Kind("HTTPRoute"),
					Namespace: gatewayapi.Namespace("route-ns"),
				}},
				To: []gatewayapi.ReferenceGrantTo{{
					Group: gatewayapi.Group(""),
					Kind:  gatewayapi.Kind("Service"),
					Name:  pointer.To(gatewayapi.ObjectName("backend")),
				}},
			},
		}
		updatedGrant := oldGrant.DeepCopy()
		updatedGrant.Spec.From = []gatewayapi.ReferenceGrantFrom{{
			Group:     gatewayapi.Group(gatewayapi.GroupVersion.Group),
			Kind:      gatewayapi.Kind("HTTPRoute"),
			Namespace: gatewayapi.Namespace("other-route-ns"),
		}}

		client := kube_client_fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(matchingRoute).
			Build()

		queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[kube_reconcile.Request]())
		DeferCleanup(queue.ShutDown)

		kube_handler.EnqueueRequestsFromMapFunc(routesForReferenceGrant(logr.Discard(), client)).Update(
			context.Background(),
			kube_event.UpdateEvent{ObjectOld: oldGrant, ObjectNew: updatedGrant},
			queue,
		)

		Expect(queue.Len()).To(Equal(1))
		req, shutdown := queue.Get()
		Expect(shutdown).To(BeFalse())
		Expect(req).To(Equal(kube_reconcile.Request{
			NamespacedName: kube_client.ObjectKeyFromObject(matchingRoute),
		}))
		queue.Done(req)
	})
})

var _ = Describe("servicesOfRoute", func() {
	serviceRef := func(namespace, name string) *gatewayapi.BackendObjectReference {
		group := gatewayapi.Group(kube_core.GroupName)
		kind := gatewayapi.Kind("Service")
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
			Name: "my-route", Namespace: "kuma-demo",
			Spec: gatewayapi.HTTPRouteSpec{
				CommonRouteSpec: gatewayapi.CommonRouteSpec{
					ParentRefs: []gatewayapi.ParentReference{
						parentRef(serviceRef("", "parent-svc")),
					},
				},
				Rules: []gatewayapi.HTTPRouteRule{
					{
						BackendRefs: []gatewayapi.HTTPBackendRef{
							{BackendObjectReference: *serviceRef("other-ns", "backend-svc")},
						},
						Filters: []gatewayapi.HTTPRouteFilter{
							{
								Type: gatewayapi_v1.HTTPRouteFilterRequestMirror,
								RequestMirror: &gatewayapi.HTTPRequestMirrorFilter{
									BackendRef: *serviceRef("mirror-ns", "mirror-svc"),
								},
							},
						},
					},
				},
			},
		}

		names := servicesOfRoute(route)
		Expect(names).To(ConsistOf(
			"kuma-demo/parent-svc",
			"other-ns/backend-svc",
			"mirror-ns/mirror-svc",
		))
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
			Name: "my-route", Namespace: "kuma-demo",
			Spec: gatewayapi.HTTPRouteSpec{
				CommonRouteSpec: gatewayapi.CommonRouteSpec{
					ParentRefs: []gatewayapi.ParentReference{
						parentRef(meshServiceRef("", "parent-ms")),
					},
				},
				Rules: []gatewayapi.HTTPRouteRule{
					{
						BackendRefs: []gatewayapi.HTTPBackendRef{
							{BackendObjectReference: *meshServiceRef("other-ns", "backend-ms")},
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
