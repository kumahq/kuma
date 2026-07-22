package gatewayapi

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	bootstrap_k8s "github.com/kumahq/kuma/v3/pkg/plugins/bootstrap/k8s"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
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
			Name: pointer.To("backend_kuma-demo_svc_80"),
		}))
	})
})
