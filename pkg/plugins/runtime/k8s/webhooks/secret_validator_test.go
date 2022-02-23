package webhooks_test

import (
	"context"
	"fmt"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	secrets_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	core_validators "github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
)

var _ = Describe("ServiceValidator", func() {

	type testCase struct {
		request  string
		expected string
	}

	BeforeEach(func() {
		secret := &kube_core.Secret{
			ObjectMeta: kube_meta.ObjectMeta{
				Name:      "secret-in-use",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"value": []byte("dGVzdAo="),
			},
			Type: "system.kuma.io/secret",
		}
		err := k8sClient.Create(context.Background(), secret)
		Expect(err).ToNot(HaveOccurred())

		secret = &kube_core.Secret{
			ObjectMeta: kube_meta.ObjectMeta{
				Name:      "secret-not-in-use",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"value": []byte("dGVzdAo="),
			},
			Type: "system.kuma.io/secret",
		}
		err = k8sClient.Create(context.Background(), secret)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := k8sClient.DeleteAllOf(context.Background(), &kube_core.Secret{}, kube_client.InNamespace("default"))
		Expect(err).ToNot(HaveOccurred())
	})

	DescribeTable("should make a proper admission verdict",
		func(given testCase) {
			// given
			validator := &webhooks.SecretValidator{
				Decoder:   decoder,
				Client:    k8sClient,
				Validator: &testSecretValidator{},
			}
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
		Entry("should allow properly constructed mesh Secret", testCase{
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
		Entry("should allow properly constructed global Secret", testCase{
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
                  name: global-sec-1
                  namespace: kuma-system
                data:
                  value: dGVzdAo=
                type: system.kuma.io/global-secret
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
		Entry("should allow Secret with mesh that does not exist", testCase{
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
              allowed: true
              status:
                code: 200
                metadata: {}
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
		Entry("should not allow mesh in global Secret", testCase{
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
                type: system.kuma.io/global-secret
              operation: CREATE
`,
			expected: `
            allowed: false
            status:
              code: 422
              details:
                causes:
                - field: metadata.labels["kuma.io/mesh"]
                  message: mesh cannot be set on global secret
                  reason: FieldValueInvalid
                kind: Secret
                name: sec-1
              message: 'metadata.labels["kuma.io/mesh"]: mesh cannot be set on global secret'
              metadata: {}
              reason: Invalid
              status: Failure
            uid: ""`,
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
		Entry("should not allow global Secret without data", testCase{
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
                type: system.kuma.io/global-secret
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
		Entry("should not allow deleting secret in use", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: Secret
                version: v1
              name: secret-in-use
              namespace: default
              operation: DELETE
`,
			expected: `
            allowed: false
            status:
              code: 422
              details:
                causes:
                - field: name
                  message: The secret "secret-in-use" that you are trying to remove is currently
                    in use in Mesh "default" in mTLS backend "ca-1". Please remove the reference
                    from the "ca-1" backend before removing the secret.
                  reason: FieldValueInvalid
                name: secret-in-use
              message: 'name: The secret "secret-in-use" that you are trying to remove is currently
                in use in Mesh "default" in mTLS backend "ca-1". Please remove the reference from
                the "ca-1" backend before removing the secret.'
              metadata: {}
              reason: Invalid
              status: Failure
            uid: ""`,
		}),
		Entry("should not allow deleting secret in use", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: Secret
                version: v1
              name: secret-in-use
              namespace: default
              operation: DELETE
`,
			expected: `
            allowed: false
            status:
              code: 422
              details:
                causes:
                - field: name
                  message: The secret "secret-in-use" that you are trying to remove is currently
                    in use in Mesh "default" in mTLS backend "ca-1". Please remove the reference
                    from the "ca-1" backend before removing the secret.
                  reason: FieldValueInvalid
                name: secret-in-use
              message: 'name: The secret "secret-in-use" that you are trying to remove is currently
                in use in Mesh "default" in mTLS backend "ca-1". Please remove the reference from
                the "ca-1" backend before removing the secret.'
              metadata: {}
              reason: Invalid
              status: Failure
            uid: ""`,
		}),
		Entry("should not allow deleting secret in use", testCase{
			request: `
            apiVersion: admission.k8s.io/v1
            kind: AdmissionReview
            request:
              uid: 12345
              kind:
                group: ""
                kind: Secret
                version: v1
              name: secret-not-in-use
              namespace: default
              operation: DELETE
`,
			expected: `
            allowed: true
            status:
              code: 200
              metadata: {}
            uid: ""`,
		}),
	)
})

type testSecretValidator struct {
}

func (t *testSecretValidator) ValidateDelete(ctx context.Context, secretName string, secretMesh string) error {
	var verr core_validators.ValidationError
	if secretName == "secret-in-use" {
		verr.AddViolation("name", fmt.Sprintf(`The secret %q that you are trying to remove is currently in use in Mesh %q in mTLS backend %q. Please remove the reference from the %q backend before removing the secret.`, secretName, secretMesh, "ca-1", "ca-1"))
	}
	return verr.OrNil()
}

var _ secrets_manager.SecretValidator = &testSecretValidator{}
