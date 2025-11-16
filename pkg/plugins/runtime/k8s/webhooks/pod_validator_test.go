package webhooks_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/webhooks"
)

var _ = Describe("PodValidator", func() {
	type testCase struct {
		request  string
		expected string
	}

	DescribeTable("should make a proper admission verdict",
		func(given testCase) {
			// given
			validator := webhooks.NewPodValidatorWebhook(decoder)
			admissionReview := admissionv1.AdmissionReview{}
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
		Entry("should allow create Pod without kuma.io/workload label", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: "Pod"
                version: v1
              name: test-pod
              namespace: default
              object:
                apiVersion: v1
                kind: Pod
                metadata:
                  name: test-pod
                  namespace: default
                  labels:
                    app: test
                spec:
                  containers:
                  - name: app
                    image: nginx
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
		Entry("should deny create Pod with kuma.io/workload label", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: "Pod"
                version: v1
              name: test-pod
              namespace: default
              object:
                apiVersion: v1
                kind: Pod
                metadata:
                  name: test-pod
                  namespace: default
                  labels:
                    app: test
                    kuma.io/workload: manually-set
                spec:
                  containers:
                  - name: app
                    image: nginx
              operation: CREATE
`,
			expected: `
            allowed: false
            status:
              code: 403
              message: "cannot manually set kuma.io/workload label on Pod; it is automatically managed by Kuma"
              metadata: {}
              reason: Forbidden
            uid: ""
`,
		}),
		Entry("should allow Pod with other kuma.io labels", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: "Pod"
                version: v1
              name: test-pod
              namespace: default
              object:
                apiVersion: v1
                kind: Pod
                metadata:
                  name: test-pod
                  namespace: default
                  labels:
                    app: test
                    kuma.io/mesh: default
                    kuma.io/sidecar-injection: enabled
                spec:
                  containers:
                  - name: app
                    image: nginx
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
		Entry("should allow Pod deletion", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: "Pod"
                version: v1
              name: test-pod
              namespace: default
              oldObject:
                apiVersion: v1
                kind: Pod
                metadata:
                  name: test-pod
                  namespace: default
                  labels:
                    app: test
                    kuma.io/workload: manually-set
                spec:
                  containers:
                  - name: app
                    image: nginx
              operation: DELETE
`,
			expected: `
            allowed: true
            status:
              code: 200
              metadata: {}
            uid: ""
`,
		}),
	)
})
