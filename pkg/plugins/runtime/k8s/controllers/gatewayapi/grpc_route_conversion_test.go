package gatewayapi

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi_beta "sigs.k8s.io/gateway-api/apis/v1beta1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshservice_k8s "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/k8s/v1alpha1"
	bootstrap_k8s "github.com/kumahq/kuma/v3/pkg/plugins/bootstrap/k8s"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var _ = Describe("gapiGRPCToKumaMeshMatch", func() {
	It("maps an exact gRPC service match to an HTTP/2 path prefix", func() {
		reconciler := &GRPCRouteReconciler{}

		match, ok := reconciler.gapiGRPCToKumaMeshMatch(gatewayapi.GRPCRouteMatch{
			Method: &gatewayapi.GRPCMethodMatch{
				Type:    pointer.To(gatewayapi.GRPCMethodMatchExact),
				Service: pointer.To("acme.echo.Echo"),
			},
		})

		Expect(ok).To(BeTrue())
		Expect(match.Path).To(Equal(&meshhttproute_api.PathMatch{
			Type:  meshhttproute_api.PathPrefix,
			Value: "/acme.echo.Echo/",
		}))
	})

	It("maps an exact gRPC service and method match to an exact HTTP/2 path", func() {
		reconciler := &GRPCRouteReconciler{}

		match, ok := reconciler.gapiGRPCToKumaMeshMatch(gatewayapi.GRPCRouteMatch{
			Method: &gatewayapi.GRPCMethodMatch{
				Type:    pointer.To(gatewayapi.GRPCMethodMatchExact),
				Service: pointer.To("acme.echo.Echo"),
				Method:  pointer.To("EchoTwo"),
			},
		})

		Expect(ok).To(BeTrue())
		Expect(match.Path).To(Equal(&meshhttproute_api.PathMatch{
			Type:  meshhttproute_api.Exact,
			Value: "/acme.echo.Echo/EchoTwo",
		}))
	})
})

var _ = Describe("gapiGRPCToKumaMeshRule", func() {
	serviceBackendRef := func(name string, weight int32) gatewayapi.GRPCBackendRef {
		group := gatewayapi.Group("")
		kind := gatewayapi.Kind("Service")
		port := gatewayapi.PortNumber(80)
		return gatewayapi.GRPCBackendRef{
			BackendRef: gatewayapi.BackendRef{
				BackendObjectReference: gatewayapi.BackendObjectReference{
					Group: &group,
					Kind:  &kind,
					Name:  gatewayapi.ObjectName(name),
					Port:  &port,
				},
				Weight: pointer.To(weight),
			},
		}
	}

	It("converts weighted backend refs and supported filters", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		backendA := &kube_core.Service{Name: "backend-a", Namespace: "kuma-demo", Spec: kube_core.ServiceSpec{Ports: []kube_core.ServicePort{{Name: "grpc", Port: 80}}}}
		backendB := &kube_core.Service{Name: "backend-b", Namespace: "kuma-demo", Spec: kube_core.ServiceSpec{Ports: []kube_core.ServicePort{{Name: "grpc", Port: 80}}}}
		mirror := &meshservice_k8s.MeshService{Name: "mirror", Namespace: "kuma-demo", Spec: &meshservice_api.MeshService{Ports: []meshservice_api.Port{{Port: 80, Name: pointer.To("grpc")}}}}
		reconciler := &GRPCRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).WithObjects(backendA, backendB, mirror).Build(),
		}

		group := gatewayapi.Group(meshservice_k8s.GroupVersion.Group)
		kind := gatewayapi.Kind("MeshService")
		ns := gatewayapi.Namespace("kuma-demo")

		rule, conditions, err := reconciler.gapiGRPCToKumaMeshRule(context.Background(), &gatewayapi.GRPCRoute{
			ObjectMeta: kube_meta.ObjectMeta{Name: "route", Namespace: "kuma-demo"},
		}, gatewayapi.GRPCRouteRule{
			Matches: []gatewayapi.GRPCRouteMatch{{
				Method: &gatewayapi.GRPCMethodMatch{
					Type:    pointer.To(gatewayapi.GRPCMethodMatchExact),
					Service: pointer.To("acme.echo.Echo"),
				},
			}},
			Filters: []gatewayapi.GRPCRouteFilter{
				{
					Type: gatewayapi.GRPCRouteFilterRequestHeaderModifier,
					RequestHeaderModifier: &gatewayapi.HTTPHeaderFilter{
						Set: []gatewayapi.HTTPHeader{{Name: "x-route", Value: "grpc"}},
					},
				},
				{
					Type: gatewayapi.GRPCRouteFilterResponseHeaderModifier,
					ResponseHeaderModifier: &gatewayapi.HTTPHeaderFilter{
						Add: []gatewayapi.HTTPHeader{{Name: "x-response", Value: "ok"}},
					},
				},
				{
					Type: gatewayapi.GRPCRouteFilterRequestMirror,
					RequestMirror: &gatewayapi.HTTPRequestMirrorFilter{
						BackendRef: gatewayapi.BackendObjectReference{
							Group:     &group,
							Kind:      &kind,
							Namespace: &ns,
							Name:      "mirror",
							Port:      pointer.To(gatewayapi.PortNumber(80)),
						},
					},
				},
			},
			BackendRefs: []gatewayapi.GRPCBackendRef{
				serviceBackendRef("backend-a", 80),
				serviceBackendRef("backend-b", 20),
			},
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(conditions).To(BeEmpty())
		Expect(rule.Default.BackendRefs).ToNot(BeNil())
		Expect(*rule.Default.BackendRefs).To(HaveLen(2))
		Expect(pointer.Deref((*rule.Default.BackendRefs)[0].Weight)).To(Equal(uint(80)))
		Expect(pointer.Deref((*rule.Default.BackendRefs)[1].Weight)).To(Equal(uint(20)))
		Expect(*rule.Default.Filters).To(HaveLen(3))
		Expect((*rule.Default.Filters)[0].Type).To(Equal(meshhttproute_api.RequestHeaderModifierType))
		Expect((*rule.Default.Filters)[1].Type).To(Equal(meshhttproute_api.ResponseHeaderModifierType))
		Expect((*rule.Default.Filters)[2].Type).To(Equal(meshhttproute_api.RequestMirrorType))
	})

	It("reports Accepted=False for unsupported extension filters", func() {
		reconciler := &GRPCRouteReconciler{}
		extensionGroup := gatewayapi.Group("example.com")
		extensionKind := gatewayapi.Kind("AuthFilter")

		rule, conditions, err := reconciler.gapiGRPCToKumaMeshRule(context.Background(), &gatewayapi.GRPCRoute{
			ObjectMeta: kube_meta.ObjectMeta{Name: "route", Namespace: "kuma-demo"},
		}, gatewayapi.GRPCRouteRule{
			Filters: []gatewayapi.GRPCRouteFilter{{
				Type: gatewayapi.GRPCRouteFilterExtensionRef,
				ExtensionRef: &gatewayapi.LocalObjectReference{
					Group: extensionGroup,
					Kind:  extensionKind,
					Name:  "authn",
				},
			}},
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(*rule.Default.Filters).To(BeEmpty())
		accepted := kube_apimeta.FindStatusCondition(conditions, string(gatewayapi.RouteConditionAccepted))
		Expect(accepted).ToNot(BeNil())
		Expect(accepted.Status).To(Equal(kube_meta.ConditionFalse))
		Expect(accepted.Reason).To(Equal(string(gatewayapi.RouteReasonUnsupportedValue)))
	})

	It("requires a ReferenceGrant whose from.kind is GRPCRoute", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		svc := &kube_core.Service{Name: "backend", Namespace: "backend-ns", Spec: kube_core.ServiceSpec{Ports: []kube_core.ServicePort{{Name: "grpc", Port: 80}}}}
		httpGrant := &gatewayapi_beta.ReferenceGrant{
			ObjectMeta: kube_meta.ObjectMeta{Name: "http-allow", Namespace: "backend-ns"},
			Spec: gatewayapi_beta.ReferenceGrantSpec{
				From: []gatewayapi_beta.ReferenceGrantFrom{{
					Group:     gatewayapi_beta.Group(gatewayapi.GroupVersion.Group),
					Kind:      gatewayapi_beta.Kind(sourceRouteKindHTTPRoute),
					Namespace: gatewayapi_beta.Namespace("route-ns"),
				}},
				To: []gatewayapi_beta.ReferenceGrantTo{{
					Group: gatewayapi_beta.Group(""),
					Kind:  gatewayapi_beta.Kind("Service"),
					Name:  pointer.To(gatewayapi_beta.ObjectName("backend")),
				}},
			},
		}
		grpcGrant := httpGrant.DeepCopy()
		grpcGrant.Name = "grpc-allow"
		grpcGrant.Spec.From[0].Kind = gatewayapi_beta.Kind(sourceRouteKindGRPCRoute)

		reconciler := &GRPCRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc, httpGrant, grpcGrant).Build(),
		}

		group := gatewayapi.Group("")
		kind := gatewayapi.Kind("Service")
		ns := gatewayapi.Namespace("backend-ns")
		ref := gatewayapi.BackendObjectReference{
			Group:     &group,
			Kind:      &kind,
			Namespace: &ns,
			Name:      "backend",
			Port:      pointer.To(gatewayapi.PortNumber(80)),
		}

		targetRef, condition, err := reconciler.uncheckedGapiToKumaRef(context.Background(), sourceRouteKindGRPCRoute, "route-ns", ref)
		Expect(err).ToNot(HaveOccurred())
		Expect(condition).To(BeNil())
		Expect(targetRef.Kind).To(Equal(common_api.MeshService))

		reconciler.Client = kube_client_fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc, httpGrant).Build()
		targetRef, condition, err = reconciler.uncheckedGapiToKumaRef(context.Background(), sourceRouteKindGRPCRoute, "route-ns", ref)
		Expect(err).ToNot(HaveOccurred())
		Expect(condition).ToNot(BeNil())
		Expect(condition.Reason).To(Equal(string(gatewayapi.RouteReasonRefNotPermitted)))
		Expect(targetRef).To(BeZero())
	})
})
