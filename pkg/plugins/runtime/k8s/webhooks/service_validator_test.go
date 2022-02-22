package webhooks_test

import (
	"context"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s"
	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
)

var _ = Describe("ServiceValidator", func() {

	var decoder *kube_admission.Decoder

	BeforeEach(func() {
		scheme, err := k8s.NewScheme()
		// expect
		Expect(err).ToNot(HaveOccurred())

		// when
		decoder, err = kube_admission.NewDecoder(scheme)
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	type testCase struct {
		request  string
		expected string
	}

	DescribeTable("should make a proper admission verdict",
		func(given testCase) {
			// setup
			validator := &ServiceValidator{}
			// when
			err := validator.InjectDecoder(decoder)
			// then
			Expect(err).ToNot(HaveOccurred())

			// setup
			admissionReview := admissionv1.AdmissionReview{}
			// when
			err = yaml.Unmarshal([]byte(given.request), &admissionReview)
			// then
			Expect(err).ToNot(HaveOccurred())

			// do
			resp := validator.Handle(context.Background(), kube_admission.Request{
				AdmissionRequest: *admissionReview.Request,
			})

			// when
			actual, err := yaml.Marshal(resp.AdmissionResponse)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("Service w/o Kuma-specific annotation", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: Service
                version: v1
              name: backend
              namespace: kuma-example
              object:
                apiVersion: v1
                kind: Service
                spec:
                  ports:
                  - port: 8080
                    targetPort: 8080
              operation: UPDATE
`,
			expected: `
            allowed: true
            status:
              code: 200
              metadata: {}
            uid: ""
`,
		}),
		Entry("Service w/ valid `<port>.service.kuma.io/protocol` annotations", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: Service
                version: v1
              name: backend
              namespace: kuma-example
              object:
                apiVersion: v1
                kind: Service
                metadata:
                  annotations:
                    8080.service.kuma.io/protocol: http
                    5432.service.kuma.io/protocol: tcp
                    1234.service.kuma.io/protocol: invalid-value # should be ignored unless this Service actually declares port '1234'
                spec:
                  ports:
                  - port: 8080
                    targetPort: 8080
                  - port: 5432
                    targetPort: 5432
              operation: UPDATE
`,
			expected: `
            allowed: true
            status:
              code: 200
              metadata: {}
            uid: ""
`,
		}),
		Entry("Service w/ multiple invalid `<port>.service.kuma.io/protocol` annotations", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: Service
                version: v1
              name: backend
              namespace: kuma-example
              object:
                apiVersion: v1
                kind: Service
                metadata:
                  annotations:
                    8080.service.kuma.io/protocol: http                       # valid protocol
                    8081.service.kuma.io/protocol: ""                         # invalid empty value
                    8082.service.kuma.io/protocol: not-yet-supported-protocol # invalid unknown value
                spec:
                  ports:
                  - port: 8080
                    targetPort: 8080
                  - port: 8081
                    targetPort: 8081
                  - port: 8082
                    targetPort: 8082
              operation: UPDATE
`,
			expected: `
            allowed: false
            status:
              code: 422
              details:
                causes:
                - field: metadata.annotations["8081.service.kuma.io/protocol"]
                  message: 'value "" is not valid. Allowed values: grpc, http, http2, kafka, tcp'
                  reason: FieldValueInvalid
                - field: metadata.annotations["8082.service.kuma.io/protocol"]
                  message: 'value "not-yet-supported-protocol" is not valid. Allowed values: grpc, http, http2, kafka, tcp'
                  reason: FieldValueInvalid
                kind: Service
              message: 'metadata.annotations["8081.service.kuma.io/protocol"]: value "" is
                not valid. Allowed values: grpc, http, http2, kafka, tcp; metadata.annotations["8082.service.kuma.io/protocol"]:
                value "not-yet-supported-protocol" is not valid. Allowed values: grpc, http, http2, kafka, tcp'
              metadata: {}
              reason: Invalid
              status: Failure
            uid: ""
`,
		}),
	)
})
