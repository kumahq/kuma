package webhooks_test

import (
	"context"
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admission "k8s.io/api/admission/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
)

var _ = Describe("Gateway API mutlizone validation webhook", func() {
	type testCase struct {
		cpMode       config_core.CpMode
		gatewayClass gatewayapi.GatewayClass
		response     kube_admission.Response
	}

	KumaGatewayClass := gatewayapi.GatewayClass{
		Spec: gatewayapi.GatewayClassSpec{
			ControllerName: common.ControllerName,
		},
	}

	DescribeTable("should validate",
		func(given testCase) {
			// given
			validator := webhooks.GatewayAPIMultizoneValidator{
				CpMode:  given.cpMode,
				Decoder: decoder,
			}

			bytes, err := json.Marshal(&given.gatewayClass)
			Expect(err).ToNot(HaveOccurred())

			req := kube_admission.Request{
				AdmissionRequest: admission.AdmissionRequest{
					Operation: admission.Create,
					UID:       "12345",
					Object: kube_runtime.RawExtension{
						Raw: bytes,
					},
					Kind: kube_meta.GroupVersionKind{
						Group:   given.gatewayClass.GetObjectKind().GroupVersionKind().Group,
						Version: given.gatewayClass.GetObjectKind().GroupVersionKind().Version,
						Kind:    given.gatewayClass.GetObjectKind().GroupVersionKind().Kind,
					},
				},
			}

			// when
			resp := validator.Handle(context.Background(), req)

			// then
			Expect(resp).To(Equal(given.response))
		},
		Entry("pass on standalone with Kuma gateway", testCase{
			cpMode:       config_core.Standalone,
			gatewayClass: KumaGatewayClass,
			response:     kube_admission.Allowed(""),
		}),
		Entry("do not pass on Zone with Kuma gateway", testCase{
			cpMode:       config_core.Zone,
			gatewayClass: KumaGatewayClass,
			response:     kube_admission.Errored(http.StatusBadRequest, webhooks.GatewayAPINotSupportedErr),
		}),
		Entry("do not pass on Global with Kuma gateway", testCase{
			cpMode:       config_core.Global,
			gatewayClass: KumaGatewayClass,
			response:     kube_admission.Errored(http.StatusBadRequest, webhooks.GatewayAPINotSupportedErr),
		}),
		Entry("pass unknown gateway on Global", testCase{
			cpMode: config_core.Global,
			gatewayClass: gatewayapi.GatewayClass{
				Spec: gatewayapi.GatewayClassSpec{
					ControllerName: "some-other-controller",
				},
			},
			response: kube_admission.Allowed(""),
		}),
	)
})
