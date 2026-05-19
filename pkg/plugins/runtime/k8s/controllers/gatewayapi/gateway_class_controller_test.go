package gatewayapi

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
)

var _ = Describe("gatewayToClassMapper", func() {
	It("maps startup events to every Kuma GatewayClass", func() {
		scheme := kube_runtime.NewScheme()
		Expect(gatewayapi.Install(scheme)).To(Succeed())

		client := kube_client_fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(
				&gatewayapi.GatewayClass{
					ObjectMeta: kube_meta.ObjectMeta{Name: "kuma"},
					Spec: gatewayapi.GatewayClassSpec{
						ControllerName: common.ControllerName,
					},
				},
				&gatewayapi.GatewayClass{
					ObjectMeta: kube_meta.ObjectMeta{Name: "other"},
					Spec: gatewayapi.GatewayClassSpec{
						ControllerName: gatewayapi.GatewayController("other.example/controller"),
					},
				},
			).
			Build()

		requests := gatewayToClassMapper(logr.Discard(), client)(context.Background(), nil)

		Expect(requests).To(ConsistOf(kube_reconcile.Request{
			NamespacedName: kube_types.NamespacedName{Name: "kuma"},
		}))
	})
})
