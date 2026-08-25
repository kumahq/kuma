package gatewayapi

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshservice_k8s "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/k8s/v1alpha1"
	bootstrap_k8s "github.com/kumahq/kuma/v3/pkg/plugins/bootstrap/k8s"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var _ = Describe("uncheckedGapiToKumaRef", func() {
	serviceRef := func(namespace, name string) gatewayapi.BackendObjectReference {
		group := gatewayapi.Group("")
		kind := gatewayapi.Kind("Service")
		port := gatewayapi.PortNumber(80)
		ref := gatewayapi.BackendObjectReference{
			Group: &group,
			Kind:  &kind,
			Name:  gatewayapi.ObjectName(name),
			Port:  &port,
		}
		if namespace != "" {
			ns := gatewayapi.Namespace(namespace)
			ref.Namespace = &ns
		}
		return ref
	}

	referenceGrant := func(grantNamespace, routeNamespace, targetKind, targetName string) *gatewayapi.ReferenceGrant {
		from := gatewayapi.ReferenceGrantFrom{
			Group:     gatewayapi.Group(gatewayapi.GroupVersion.Group),
			Kind:      gatewayapi.Kind("HTTPRoute"),
			Namespace: gatewayapi.Namespace(routeNamespace),
		}
		to := gatewayapi.ReferenceGrantTo{
			Group: gatewayapi.Group(""),
			Kind:  gatewayapi.Kind(targetKind),
		}
		if targetKind == "MeshService" {
			to.Group = gatewayapi.Group(meshservice_k8s.GroupVersion.Group)
		}
		if targetName != "" {
			to.Name = pointer.To(gatewayapi.ObjectName(targetName))
		}
		return &gatewayapi.ReferenceGrant{
			Name:      "allow-route",
			Namespace: grantNamespace,
			Spec: gatewayapi.ReferenceGrantSpec{
				From: []gatewayapi.ReferenceGrantFrom{from},
				To:   []gatewayapi.ReferenceGrantTo{to},
			},
		}
	}

	It("should convert a Service backendRef to a MeshService targetRef without tags", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		svc := &kube_core.Service{
			Name:      "backend",
			Namespace: "kuma-demo",
			Spec: kube_core.ServiceSpec{
				Ports: []kube_core.ServicePort{{Port: 80}},
			},
		}

		reconciler := &HTTPRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build(),
			Zone:   "zone-1",
		}

		ref := serviceRef("kuma-demo", "backend")

		targetRef, condition, err := reconciler.uncheckedGapiToKumaRef(context.Background(), "kuma-demo", ref)

		Expect(err).ToNot(HaveOccurred())
		Expect(condition).To(BeNil())
		Expect(targetRef).To(Equal(common_api.TargetRef{
			Kind: common_api.MeshService,
			Labels: pointer.To(map[string]string{
				mesh_proto.DisplayName:      "backend",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			}),
			SectionName: pointer.To("80"),
		}))
	})

	It("should report ResolvedRefs=False when the referenced Service does not have the requested port", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		svc := &kube_core.Service{
			Name:      "backend",
			Namespace: "kuma-demo",
			Spec: kube_core.ServiceSpec{
				Ports: []kube_core.ServicePort{{Port: 8080}},
			},
		}

		reconciler := &HTTPRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build(),
			Zone:   "zone-1",
		}

		ref := serviceRef("kuma-demo", "backend")

		targetRef, condition, err := reconciler.uncheckedGapiToKumaRef(context.Background(), "kuma-demo", ref)

		Expect(err).ToNot(HaveOccurred())
		Expect(condition).ToNot(BeNil())
		Expect(condition.Reason).To(Equal(string(gatewayapi.RouteReasonBackendNotFound)))
		Expect(condition.Message).To(ContainSubstring("kuma-demo/backend"))
		Expect(condition.Message).To(ContainSubstring("80"))
		Expect(targetRef).To(Equal(common_api.TargetRef{
			Kind: common_api.MeshService,
			Labels: pointer.To(map[string]string{
				mesh_proto.DisplayName:      "backend",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			}),
			SectionName: pointer.To("80"),
		}))
	})

	meshServiceRef := func(port *gatewayapi.PortNumber) gatewayapi.BackendObjectReference {
		group := gatewayapi.Group(meshservice_k8s.GroupVersion.Group)
		kind := gatewayapi.Kind("MeshService")
		ns := gatewayapi.Namespace("kuma-demo")
		return gatewayapi.BackendObjectReference{
			Group:     &group,
			Kind:      &kind,
			Name:      "backend",
			Namespace: &ns,
			Port:      port,
		}
	}

	It("should convert a MeshService backendRef without a port to a targetRef with no sectionName", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		ms := &meshservice_k8s.MeshService{
			Name: "backend", Namespace: "kuma-demo",
			Spec: &meshservice_api.MeshService{
				Ports: []meshservice_api.Port{{Port: 80}},
			},
		}

		reconciler := &HTTPRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).WithObjects(ms).Build(),
		}

		targetRef, condition, err := reconciler.uncheckedGapiToKumaRef(context.Background(), "kuma-demo", meshServiceRef(nil))

		Expect(err).ToNot(HaveOccurred())
		Expect(condition).To(BeNil())
		Expect(targetRef).To(Equal(common_api.TargetRef{
			Kind: common_api.MeshService,
			Labels: pointer.To(map[string]string{
				mesh_proto.DisplayName:      "backend",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			}),
		}))
	})

	It("should convert a MeshService backendRef with a matching port to a targetRef with a sectionName", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		ms := &meshservice_k8s.MeshService{
			Name: "backend", Namespace: "kuma-demo",
			Spec: &meshservice_api.MeshService{
				Ports: []meshservice_api.Port{{Port: 80, Name: pointer.To("http")}},
			},
		}

		reconciler := &HTTPRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).WithObjects(ms).Build(),
		}

		targetRef, condition, err := reconciler.uncheckedGapiToKumaRef(context.Background(), "kuma-demo", meshServiceRef(pointer.To(gatewayapi.PortNumber(80))))

		Expect(err).ToNot(HaveOccurred())
		Expect(condition).To(BeNil())
		Expect(targetRef).To(Equal(common_api.TargetRef{
			Kind: common_api.MeshService,
			Labels: pointer.To(map[string]string{
				mesh_proto.DisplayName:      "backend",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			}),
			SectionName: pointer.To("http"),
		}))
	})

	It("should report ResolvedRefs=False when the referenced MeshService does not exist", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		reconciler := &HTTPRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).Build(),
		}

		targetRef, condition, err := reconciler.uncheckedGapiToKumaRef(context.Background(), "kuma-demo", meshServiceRef(nil))

		Expect(err).ToNot(HaveOccurred())
		Expect(condition).ToNot(BeNil())
		Expect(condition.Reason).To(Equal(string(gatewayapi.RouteReasonBackendNotFound)))
		Expect(targetRef.Kind).To(Equal(common_api.MeshService))
	})

	It("should report ResolvedRefs=False when the referenced port does not exist on the MeshService", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		ms := &meshservice_k8s.MeshService{
			Name: "backend", Namespace: "kuma-demo",
			Spec: &meshservice_api.MeshService{
				Ports: []meshservice_api.Port{{Port: 80}},
			},
		}

		reconciler := &HTTPRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).WithObjects(ms).Build(),
		}

		targetRef, condition, err := reconciler.uncheckedGapiToKumaRef(context.Background(), "kuma-demo", meshServiceRef(pointer.To(gatewayapi.PortNumber(443))))

		Expect(err).ToNot(HaveOccurred())
		Expect(condition).ToNot(BeNil())
		Expect(condition.Reason).To(Equal(string(gatewayapi.RouteReasonBackendNotFound)))
		Expect(condition.Message).To(ContainSubstring("443"))
		Expect(targetRef).To(Equal(common_api.TargetRef{
			Kind: common_api.MeshService,
			Labels: pointer.To(map[string]string{
				mesh_proto.DisplayName:      "backend",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			}),
			SectionName: pointer.To("443"),
		}))
	})

	It("should deny a cross-namespace Service backendRef without a matching ReferenceGrant", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		svc := &kube_core.Service{
			Name: "backend", Namespace: "backend-ns",
			Spec: kube_core.ServiceSpec{
				Ports: []kube_core.ServicePort{{Port: 80}},
			},
		}

		reconciler := &HTTPRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build(),
		}

		targetRef, condition, err := reconciler.uncheckedGapiToKumaRef(context.Background(), "route-ns", serviceRef("backend-ns", "backend"))

		Expect(err).ToNot(HaveOccurred())
		Expect(condition).ToNot(BeNil())
		Expect(condition.Reason).To(Equal(string(gatewayapi.RouteReasonRefNotPermitted)))
		Expect(condition.Message).To(ContainSubstring("backend-ns/backend"))
		Expect(targetRef).To(BeZero())
	})

	It("should allow a cross-namespace Service backendRef when an exact-name ReferenceGrant matches", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		svc := &kube_core.Service{
			Name: "backend", Namespace: "backend-ns",
			Spec: kube_core.ServiceSpec{
				Ports: []kube_core.ServicePort{{Name: "http", Port: 80}},
			},
		}
		grant := referenceGrant("backend-ns", "route-ns", "Service", "backend")

		reconciler := &HTTPRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc, grant).Build(),
		}

		targetRef, condition, err := reconciler.uncheckedGapiToKumaRef(context.Background(), "route-ns", serviceRef("backend-ns", "backend"))

		Expect(err).ToNot(HaveOccurred())
		Expect(condition).To(BeNil())
		Expect(targetRef.SectionName).To(Equal(pointer.To("http")))
	})

	It("should allow a cross-namespace MeshService backendRef when a wildcard-name ReferenceGrant matches", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		ms := &meshservice_k8s.MeshService{
			Name: "backend", Namespace: "backend-ns",
			Spec: &meshservice_api.MeshService{
				Ports: []meshservice_api.Port{{Port: 80, Name: pointer.To("http")}},
			},
		}
		grant := referenceGrant("backend-ns", "route-ns", "MeshService", "")

		reconciler := &HTTPRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).WithObjects(ms, grant).Build(),
		}

		targetRef, condition, err := reconciler.uncheckedGapiToKumaRef(context.Background(), "route-ns", func() gatewayapi.BackendObjectReference {
			ref := meshServiceRef(pointer.To(gatewayapi.PortNumber(80)))
			ns := gatewayapi.Namespace("backend-ns")
			ref.Namespace = &ns
			return ref
		}())

		Expect(err).ToNot(HaveOccurred())
		Expect(condition).To(BeNil())
		Expect(targetRef.SectionName).To(Equal(pointer.To("http")))
	})

	It("should deny a cross-namespace Service backendRef when the ReferenceGrant targets a different backend name", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		svc := &kube_core.Service{
			Name: "backend", Namespace: "backend-ns",
			Spec: kube_core.ServiceSpec{
				Ports: []kube_core.ServicePort{{Port: 80}},
			},
		}
		grant := referenceGrant("backend-ns", "route-ns", "Service", "other-backend")

		reconciler := &HTTPRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc, grant).Build(),
		}

		_, condition, err := reconciler.uncheckedGapiToKumaRef(context.Background(), "route-ns", serviceRef("backend-ns", "backend"))

		Expect(err).ToNot(HaveOccurred())
		Expect(condition).ToNot(BeNil())
		Expect(condition.Reason).To(Equal(string(gatewayapi.RouteReasonRefNotPermitted)))
	})

	It("should return a valid unresolved MeshService targetRef when the backend Service is missing", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		reconciler := &HTTPRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).Build(),
			Zone:   "zone-1",
		}

		ref := serviceRef("kuma-demo", "missing-backend")

		targetRef, condition, err := reconciler.uncheckedGapiToKumaRef(context.Background(), "kuma-demo", ref)

		Expect(err).ToNot(HaveOccurred())
		Expect(condition).ToNot(BeNil())
		Expect(targetRef).To(Equal(common_api.TargetRef{
			Kind: common_api.MeshService,
			Labels: pointer.To(map[string]string{
				mesh_proto.DisplayName:      "missing-backend",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			}),
		}))
	})
})

var _ = Describe("gapiMeshServiceToMeshRoute", func() {
	meshService := func(ports ...meshservice_api.Port) *meshservice_k8s.MeshService {
		return &meshservice_k8s.MeshService{
			Name:      "backend",
			Namespace: "kuma-demo",
			Spec:      &meshservice_api.MeshService{Ports: ports},
		}
	}

	toEntries := func(spec any) []v1alpha1.To {
		route, ok := spec.(*v1alpha1.MeshHTTPRoute)
		Expect(ok).To(BeTrue())
		return pointer.Deref(route.To)
	}

	reconciler := &HTTPRouteReconciler{}

	It("produces a producer route (Mesh targetRef) when the route shares the MeshService namespace", func() {
		spec, matched := reconciler.gapiMeshServiceToMeshRoute("kuma-demo", nil,
			meshService(meshservice_api.Port{Port: 80, Name: pointer.To("http")}), nil, nil)
		Expect(matched).To(BeTrue())

		route, ok := spec.(*v1alpha1.MeshHTTPRoute)
		Expect(ok).To(BeTrue())
		Expect(route.TargetRef.Kind).To(Equal(common_api.TopLevelTargetRefKindMesh))
	})

	It("produces a consumer route (Dataplane targetRef) when the route is in a different namespace", func() {
		spec, matched := reconciler.gapiMeshServiceToMeshRoute("other-ns", nil,
			meshService(meshservice_api.Port{Port: 80, Name: pointer.To("http")}), nil, nil)
		Expect(matched).To(BeTrue())

		route, ok := spec.(*v1alpha1.MeshHTTPRoute)
		Expect(ok).To(BeTrue())
		Expect(route.TargetRef.Kind).To(Equal(common_api.TopLevelTargetRefKindDataplane))
		Expect(route.TargetRef.Labels).To(Equal(&map[string]string{
			mesh_proto.KubeNamespaceTag: "other-ns",
		}))
	})

	It("creates one `to` entry per MeshService port with the port name as sectionName", func() {
		spec, matched := reconciler.gapiMeshServiceToMeshRoute("kuma-demo", nil,
			meshService(meshservice_api.Port{Port: 80, Name: pointer.To("http")},
				meshservice_api.Port{Port: 443, Name: pointer.To("https")},
			), nil, nil)
		Expect(matched).To(BeTrue())

		entries := toEntries(spec)
		Expect(entries).To(HaveLen(2))
		Expect(pointer.Deref(entries[0].TargetRef.SectionName)).To(Equal("http"))
		Expect(pointer.Deref(entries[1].TargetRef.SectionName)).To(Equal("https"))
		Expect(entries[0].TargetRef.Kind).To(Equal(common_api.OutboundTargetRefKindMeshService))
		Expect(entries[0].TargetRef.Labels).To(Equal(&map[string]string{
			mesh_proto.DisplayName:      "backend",
			mesh_proto.KubeNamespaceTag: "kuma-demo",
		}))
	})

	It("selects only the port referenced by parentPort", func() {
		port := gatewayapi_v1.PortNumber(443)
		spec, matched := reconciler.gapiMeshServiceToMeshRoute("kuma-demo", nil,
			meshService(meshservice_api.Port{Port: 80, Name: pointer.To("http")},
				meshservice_api.Port{Port: 443, Name: pointer.To("https")},
			), &port, nil)
		Expect(matched).To(BeTrue())

		entries := toEntries(spec)
		Expect(entries).To(HaveLen(1))
		Expect(pointer.Deref(entries[0].TargetRef.SectionName)).To(Equal("https"))
	})

	It("selects only the port named by parentSectionName", func() {
		sectionName := gatewayapi_v1.SectionName("https")
		spec, matched := reconciler.gapiMeshServiceToMeshRoute("kuma-demo", nil,
			meshService(meshservice_api.Port{Port: 80, Name: pointer.To("http")},
				meshservice_api.Port{Port: 443, Name: pointer.To("https")},
			), nil, &sectionName)
		Expect(matched).To(BeTrue())

		entries := toEntries(spec)
		Expect(entries).To(HaveLen(1))
		Expect(pointer.Deref(entries[0].TargetRef.SectionName)).To(Equal("https"))
	})

	It("names an unnamed port by its number", func() {
		sectionName := gatewayapi_v1.SectionName("443")
		spec, matched := reconciler.gapiMeshServiceToMeshRoute("kuma-demo", nil,
			meshService(meshservice_api.Port{Port: 443}), nil, &sectionName)
		Expect(matched).To(BeTrue())

		entries := toEntries(spec)
		Expect(entries).To(HaveLen(1))
		Expect(pointer.Deref(entries[0].TargetRef.SectionName)).To(Equal("443"))
	})

	It("reports no match when parentPort and parentSectionName disagree", func() {
		port := gatewayapi_v1.PortNumber(80)
		sectionName := gatewayapi_v1.SectionName("https")
		_, matched := reconciler.gapiMeshServiceToMeshRoute("kuma-demo", nil,
			meshService(meshservice_api.Port{Port: 80, Name: pointer.To("http")},
				meshservice_api.Port{Port: 443, Name: pointer.To("https")},
			), &port, &sectionName)

		Expect(matched).To(BeFalse())
	})

	It("reports no match when the referenced port does not exist", func() {
		port := gatewayapi_v1.PortNumber(8080)
		_, matched := reconciler.gapiMeshServiceToMeshRoute("kuma-demo", nil,
			meshService(meshservice_api.Port{Port: 80, Name: pointer.To("http")}), &port, nil)

		Expect(matched).To(BeFalse())
	})

	It("produces no `to` entries for a MeshService with no ports", func() {
		spec, matched := reconciler.gapiMeshServiceToMeshRoute("kuma-demo", nil, meshService(), nil, nil)
		Expect(matched).To(BeTrue())

		Expect(toEntries(spec)).To(BeEmpty())
	})
})

var _ = Describe("gapiServiceToMeshRoute", func() {
	service := func(ports ...kube_core.ServicePort) *kube_core.Service {
		return &kube_core.Service{
			Name:      "backend",
			Namespace: "kuma-demo",
			Spec:      kube_core.ServiceSpec{Ports: ports},
		}
	}

	sectionNames := func(spec any) []string {
		route, ok := spec.(*v1alpha1.MeshHTTPRoute)
		Expect(ok).To(BeTrue())
		var names []string
		for _, to := range pointer.Deref(route.To) {
			Expect(to.TargetRef.Labels).To(Equal(&map[string]string{
				mesh_proto.DisplayName:      "backend",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			}))
			names = append(names, pointer.Deref(to.TargetRef.SectionName))
		}
		return names
	}

	reconciler := &HTTPRouteReconciler{}

	It("uses the Service port name as the MeshService section name", func() {
		spec, matched := reconciler.gapiServiceToMeshRoute("other-ns", nil,
			service(kube_core.ServicePort{Name: "http", Port: 80}), nil, nil)

		Expect(matched).To(BeTrue())
		Expect(sectionNames(spec)).To(Equal([]string{"http"}))
	})

	It("falls back to the stringified port value for unnamed ports", func() {
		spec, matched := reconciler.gapiServiceToMeshRoute("other-ns", nil,
			service(kube_core.ServicePort{Port: 80}), nil, nil)

		Expect(matched).To(BeTrue())
		Expect(sectionNames(spec)).To(Equal([]string{"80"}))
	})

	It("selects the port referenced by parentPort", func() {
		port := gatewayapi_v1.PortNumber(80)
		spec, matched := reconciler.gapiServiceToMeshRoute("other-ns", nil,
			service(
				kube_core.ServicePort{Name: "http", Port: 80},
				kube_core.ServicePort{Name: "https", Port: 443},
			), &port, nil)

		Expect(matched).To(BeTrue())
		Expect(sectionNames(spec)).To(Equal([]string{"http"}))
	})

	It("selects the port referenced by parentSectionName", func() {
		sectionName := gatewayapi_v1.SectionName("https")
		spec, matched := reconciler.gapiServiceToMeshRoute("other-ns", nil,
			service(
				kube_core.ServicePort{Name: "http", Port: 80},
				kube_core.ServicePort{Name: "https", Port: 443},
			), nil, &sectionName)

		Expect(matched).To(BeTrue())
		Expect(sectionNames(spec)).To(Equal([]string{"https"}))
	})

	It("selects the port when parentPort and parentSectionName agree", func() {
		port := gatewayapi_v1.PortNumber(443)
		sectionName := gatewayapi_v1.SectionName("https")
		spec, matched := reconciler.gapiServiceToMeshRoute("other-ns", nil,
			service(
				kube_core.ServicePort{Name: "http", Port: 80},
				kube_core.ServicePort{Name: "https", Port: 443},
			), &port, &sectionName)

		Expect(matched).To(BeTrue())
		Expect(sectionNames(spec)).To(Equal([]string{"https"}))
	})

	It("reports no match when parentPort and parentSectionName disagree", func() {
		port := gatewayapi_v1.PortNumber(80)
		sectionName := gatewayapi_v1.SectionName("https")
		_, matched := reconciler.gapiServiceToMeshRoute("other-ns", nil,
			service(
				kube_core.ServicePort{Name: "http", Port: 80},
				kube_core.ServicePort{Name: "https", Port: 443},
			), &port, &sectionName)

		Expect(matched).To(BeFalse())
	})

	It("reports no match when parentSectionName does not exist", func() {
		sectionName := gatewayapi_v1.SectionName("grpc")
		_, matched := reconciler.gapiServiceToMeshRoute("other-ns", nil,
			service(
				kube_core.ServicePort{Name: "http", Port: 80},
				kube_core.ServicePort{Name: "https", Port: 443},
			), nil, &sectionName)

		Expect(matched).To(BeFalse())
	})

	It("matches an unnamed port by its numeric fallback sectionName", func() {
		sectionName := gatewayapi_v1.SectionName("443")
		spec, matched := reconciler.gapiServiceToMeshRoute("other-ns", nil,
			service(kube_core.ServicePort{Port: 443}), nil, &sectionName)

		Expect(matched).To(BeTrue())
		Expect(sectionNames(spec)).To(Equal([]string{"443"}))
	})
})
