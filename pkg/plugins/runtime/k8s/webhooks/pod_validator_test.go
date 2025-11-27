package webhooks_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	kube_core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"

	mesh_k8s "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/webhooks"
)

func dataplaneInDifferentMesh(namespace, mesh string) []kube_client.Object {
	return []kube_client.Object{
		&kube_core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		},
		&mesh_k8s.Dataplane{
			Mesh: mesh,
			ObjectMeta: metav1.ObjectMeta{
				Name:      "existing-dp",
				Namespace: namespace,
			},
		},
	}
}

var _ = Describe("PodValidator", func() {
	type testCase struct {
		request                            string
		expected                           string
		disallowMultipleMeshesPerNamespace bool
		existingObjects                    []kube_client.Object
	}

	DescribeTable("should make a proper admission verdict",
		func(given testCase) {
			// given
			kubeClient := kube_client_fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(given.existingObjects...).
				Build()
			validator := webhooks.NewPodValidatorWebhook(decoder, kubeClient, given.disallowMultipleMeshesPerNamespace)
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
		Entry("should allow Pod when multiple meshes exist but flag is disabled", testCase{
			disallowMultipleMeshesPerNamespace: false,
			existingObjects:                    dataplaneInDifferentMesh("default", "mesh1"),
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
                  annotations:
                    kuma.io/mesh: mesh2
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
		Entry("should deny Pod when multiple meshes would be created and flag is enabled", testCase{
			disallowMultipleMeshesPerNamespace: true,
			existingObjects:                    dataplaneInDifferentMesh("default", "mesh1"),
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
                  annotations:
                    kuma.io/mesh: mesh2
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
              message: 'pod is in mesh "mesh2" but namespace "default" already contains dataplanes in mesh "mesh1"; only one mesh per namespace is allowed when runtime.kubernetes.disallowMultipleMeshesPerNamespace is enabled'
              metadata: {}
              reason: Forbidden
            uid: ""
`,
		}),
		Entry("should allow Pod in same mesh when flag is enabled", testCase{
			disallowMultipleMeshesPerNamespace: true,
			existingObjects:                    dataplaneInDifferentMesh("default", "mesh1"),
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
                  annotations:
                    kuma.io/mesh: mesh1
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
	)
})
