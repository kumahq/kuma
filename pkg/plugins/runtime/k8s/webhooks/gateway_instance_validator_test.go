package webhooks_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	k8s_resources "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
)

var _ = Describe("MeshGatewayInstanceValidator", func() {
	type testCase struct {
		cpMode          core.CpMode
		request         string
		expectedMessage string
	}

	DescribeTable("Pass MeshGatewayInstance Webhook validation",
		func(given testCase) {
			wh, err := newGatewayInstanceValidatorWebhook(given.cpMode)
			Expect(err).NotTo(HaveOccurred())

			// setup
			admissionReview := admissionv1.AdmissionReview{}
			// when
			err = yaml.Unmarshal([]byte(given.request), &admissionReview)
			// then
			Expect(err).NotTo(HaveOccurred())

			// do
			resp := wh.Handle(context.Background(), kube_admission.Request{
				AdmissionRequest: *admissionReview.Request,
			})

			// then
			if given.expectedMessage != "" {
				Expect(resp.Result.Message).To(Equal(given.expectedMessage))
			} else {
				Expect(resp.Result.Message).To(Equal(""))
			}
		},
		Entry("should allow create MeshGatewayInstance in Zone", testCase{
			cpMode: core.Zone,
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: "kuma.io"
                kind: "MeshGatewayInstance"
                version: v1alpha1
              name: my-gateway
              object:
                apiVersion: kuma.io/v1alpha1
                kind: MeshGatewayInstance
                metadata:
                  name: my-gateway
                  labels:
                    kuma.io/mesh: default
                spec:
                  tags:
                    custom.io/gateway: my-gateway
                  replicas: 1
                  serviceType: LoadBalancer
              operation: CREATE
`,
			expectedMessage: "",
		}),
		Entry("should not allow create MeshGatewayInstance in Global", testCase{
			cpMode: core.Global,
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: "kuma.io"
                kind: "MeshGatewayInstance"
                version: v1alpha1
              name: my-gateway
              object:
                apiVersion: kuma.io/v1alpha1
                kind: MeshGatewayInstance
                metadata:
                  name: my-gateway
                  labels:
                    kuma.io/mesh: default
                spec:
                  tags:
                    custom.io/gateway: my-gateway
                  replicas: 1
                  serviceType: LoadBalancer
              operation: CREATE
`,
			expectedMessage: "Operation not allowed. Kuma resources like MeshGatewayInstance can be created only from the 'zone' control plane and not from a 'global' control plane.",
		}),

		Entry("should not allow create MeshGatewayInstance with kuma.io/service tag", testCase{
			cpMode: core.Zone,
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: "kuma.io"
                kind: "MeshGatewayInstance"
                version: v1alpha1
              name: my-gateway
              object:
                apiVersion: kuma.io/v1alpha1
                kind: MeshGatewayInstance
                metadata:
                  name: my-gateway
                  labels:
                    kuma.io/mesh: default
                spec:
                  tags:
                    kuma.io/service: edge-gateway_kuma-demo_svc
                  replicas: 1
                  serviceType: LoadBalancer
              operation: CREATE
`,
			expectedMessage: "tags: \"kuma.io/service\" must not be defined",
		}),
	)
})

func newGatewayInstanceValidatorWebhook(mode core.CpMode) (*kube_admission.Webhook, error) {
	simpleConverter := k8s_resources.NewSimpleConverter()
	store, err := k8s_resources.NewStore(kube_client_fake.NewFakeClient(), scheme, simpleConverter)
	if err != nil {
		return nil, err
	}

	handler := webhooks.NewGatewayInstanceValidatorWebhook(simpleConverter, manager.NewResourceManager(store), mode)
	handler.InjectDecoder(kube_admission.NewDecoder(scheme))
	return &kube_admission.Webhook{
		Handler: handler,
	}, nil
}
