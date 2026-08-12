package gatewayapi

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	bootstrap_k8s "github.com/kumahq/kuma/v3/pkg/plugins/bootstrap/k8s"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/gateway/metadata"
)

var _ = Describe("uncheckedGapiToKumaRef", func() {
	It("should convert a Service backendRef to a MeshService targetRef without tags", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		svc := &kube_core.Service{
			ObjectMeta: kube_meta.ObjectMeta{
				Name:      "backend",
				Namespace: "kuma-demo",
			},
		}

		reconciler := &HTTPRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build(),
			Zone:   "zone-1",
		}

		group := gatewayapi.Group("")
		kind := gatewayapi.Kind("Service")
		namespace := gatewayapi.Namespace("kuma-demo")
		port := gatewayapi.PortNumber(80)
		ref := gatewayapi.BackendObjectReference{
			Group:     &group,
			Kind:      &kind,
			Name:      "backend",
			Namespace: &namespace,
			Port:      &port,
		}

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

	It("should return a valid unresolved MeshService targetRef when the backend Service is missing", func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		reconciler := &HTTPRouteReconciler{
			Client: kube_client_fake.NewClientBuilder().WithScheme(scheme).Build(),
			Zone:   "zone-1",
		}

		group := gatewayapi.Group("")
		kind := gatewayapi.Kind("Service")
		namespace := gatewayapi.Namespace("kuma-demo")
		port := gatewayapi.PortNumber(80)
		ref := gatewayapi.BackendObjectReference{
			Group:     &group,
			Kind:      &kind,
			Name:      "missing-backend",
			Namespace: &namespace,
			Port:      &port,
		}

		targetRef, condition, err := reconciler.uncheckedGapiToKumaRef(context.Background(), "kuma-demo", ref)

		Expect(err).ToNot(HaveOccurred())
		Expect(condition).ToNot(BeNil())
		Expect(targetRef).To(Equal(common_api.TargetRef{
			Kind: common_api.MeshService,
			Labels: pointer.To(map[string]string{
				mesh_proto.DisplayName: metadata.UnresolvedBackendServiceName,
			}),
		}))
	})
})

var _ = Describe("gapiServiceToMeshRoute", func() {
	service := func(ports ...kube_core.ServicePort) *kube_core.Service {
		return &kube_core.Service{
			ObjectMeta: kube_meta.ObjectMeta{
				Name:      "backend",
				Namespace: "kuma-demo",
			},
			Spec: kube_core.ServiceSpec{Ports: ports},
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
		spec := reconciler.gapiServiceToMeshRoute("other-ns", nil,
			service(kube_core.ServicePort{Name: "http", Port: 80}), nil)

		Expect(sectionNames(spec)).To(Equal([]string{"http"}))
	})

	It("falls back to the stringified port value for unnamed ports", func() {
		spec := reconciler.gapiServiceToMeshRoute("other-ns", nil,
			service(kube_core.ServicePort{Port: 80}), nil)

		Expect(sectionNames(spec)).To(Equal([]string{"80"}))
	})

	It("selects the named port referenced by parentRef", func() {
		port := gatewayapi_v1.PortNumber(80)
		spec := reconciler.gapiServiceToMeshRoute("other-ns", nil,
			service(
				kube_core.ServicePort{Name: "http", Port: 80},
				kube_core.ServicePort{Name: "https", Port: 443},
			), &port)

		Expect(sectionNames(spec)).To(Equal([]string{"http"}))
	})
})
