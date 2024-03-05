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

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
)

var _ = Describe("Gateway API mutlizone validation webhook", func() {
	type testCase struct {
		gapiSupported bool
		gatewayClass  gatewayapi.GatewayClass
		response      kube_admission.Response
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
				GAPISupported: given.gapiSupported,
				Decoder:       decoder,
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
		Entry("pass with Kuma gateway when gapi is supported", testCase{
			gapiSupported: true,
			gatewayClass:  KumaGatewayClass,
			response:      kube_admission.Allowed(""),
		}),
		Entry("do not pass when gapi is not supported Kuma gateway", testCase{
			gapiSupported: false,
			gatewayClass:  KumaGatewayClass,
			response:      kube_admission.Errored(http.StatusBadRequest, webhooks.GatewayAPINotSupportedErr),
		}),
		Entry("pass unknown gateway when gapi is not supported", testCase{
			gapiSupported: false,
			gatewayClass: gatewayapi.GatewayClass{
				Spec: gatewayapi.GatewayClassSpec{
					ControllerName: "some-other-controller",
				},
			},
			response: kube_admission.Allowed(""),
		}),
	)
})
