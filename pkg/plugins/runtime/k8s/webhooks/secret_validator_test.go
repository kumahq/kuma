package webhooks_test

import (
	"context"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/Kong/kuma/pkg/plugins/runtime/k8s/webhooks"
)

var _ = Describe("ServiceValidator", func() {

	type testCase struct {
		request  string
		expected string
	}

	DescribeTable("should make a proper admission verdict",
		func(given testCase) {
			// given
			validator := &webhooks.SecretValidator{
				Decoder: decoder,
				Client:  k8sClient,
			}
			admissionReview := admissionv1beta1.AdmissionReview{}
			err := yaml.Unmarshal([]byte(given.request), &admissionReview)
			Expect(err).ToNot(HaveOccurred())

			// when
			resp := validator.Handle(context.Background(), kube_admission.Request{
				AdmissionRequest: *admissionReview.Request,
			})

			// then
			actual, err := yaml.Marshal(resp.AdmissionResponse)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("should allow properly constructed Secret", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: Secret
                version: v1
              name: sec-1
              namespace: kuma-system
              object:
                apiVersion: v1
                kind: Secret
                metadata:
                  name: sec-1
                  namespace: kuma-system
                  labels:
                    kuma.io/mesh: default
                data:
                  value: dGVzdAo=
                type: system.kuma.io/secret
              operation: CREATE
`,
			expected: `
            allowed: true
            status:
              code: 200
              metadata: {}
            uid: ""
`,
		}),
		Entry("should allow Secret without mesh label (defaults to 'default')", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: Secret
                version: v1
              name: sec-1
              namespace: kuma-system
              object:
                apiVersion: v1
                kind: Secret
                metadata:
                  name: sec-1
                  namespace: kuma-system
                data:
                  value: dGVzdAo=
                type: system.kuma.io/secret
              operation: CREATE
`,
			expected: `
            allowed: true
            status:
              code: 200
              metadata: {}
            uid: ""
`,
		}),
		Entry("should not allow Secret with mesh that does not exist", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: Secret
                version: v1
              name: sec-1
              namespace: kuma-system
              object:
                apiVersion: v1
                kind: Secret
                metadata:
                  name: sec-1
                  namespace: kuma-system
                  labels:
                    kuma.io/mesh: non-existent
                data:
                  value: dGVzdAo=
                type: system.kuma.io/secret
              operation: CREATE
`,
			expected: `
            allowed: false
            status:
              code: 422
              details:
                causes:
                - field: metadata.labels["kuma.io/mesh"]
                  message: mesh does not exist
                  reason: FieldValueInvalid
                kind: Secret
                name: sec-1
              message: 'metadata.labels["kuma.io/mesh"]: mesh does not exist'
              metadata: {}
              reason: Invalid
              status: Failure
            uid: ""
`,
		}),
		Entry("should not allow switching mesh in Secret", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: Secret
                version: v1
              name: sec-1
              namespace: kuma-system
              object:
                apiVersion: v1
                kind: Secret
                metadata:
                  name: sec-1
                  namespace: kuma-system
                  labels:
                    kuma.io/mesh: default
                data:
                  value: dGVzdAo=
                type: system.kuma.io/secret
              oldObject:
                apiVersion: v1
                kind: Secret
                metadata:
                  name: sec-1
                  namespace: kuma-system
                  labels:
                    kuma.io/mesh: second
                data:
                  value: dGVzdAo=
                type: system.kuma.io/secret
              operation: UPDATE
`,
			expected: `
            allowed: false
            status:
              code: 422
              details:
                causes:
                - field: metadata.labels["kuma.io/mesh"]
                  message: cannot change mesh of the Secret. Delete the Secret first and apply it
                    again.
                  reason: FieldValueInvalid
                kind: Secret
                name: sec-1
              message: 'metadata.labels["kuma.io/mesh"]: cannot change mesh of the Secret. Delete
                the Secret first and apply it again.'
              metadata: {}
              reason: Invalid
              status: Failure
            uid: ""
`,
		}),
		Entry("should not allow Secret without data", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: Secret
                version: v1
              name: sec-1
              namespace: kuma-system
              object:
                apiVersion: v1
                kind: Secret
                metadata:
                  name: sec-1
                  namespace: kuma-system
                  labels:
                    kuma.io/mesh: default
                type: system.kuma.io/secret
              operation: CREATE
`,
			expected: `
            allowed: false
            status:
              code: 422
              details:
                causes:
                - field: data.value
                  message: cannot be empty.
                  reason: FieldValueInvalid
                kind: Secret
                name: sec-1
              message: 'data.value: cannot be empty.'
              metadata: {}
              reason: Invalid
              status: Failure
            uid: ""
`,
		}),
	)
})
