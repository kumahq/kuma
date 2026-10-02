package rest_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	mlbs_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	mtp_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
)

var _ = Describe("UnmarshalStrict", func() {
	desc := mtp_api.MeshTrafficPermissionResourceTypeDescriptor

	valid := `{
		"type": "MeshTrafficPermission",
		"name": "mtp-1",
		"mesh": "default",
		"spec": {
			"targetRef": {"kind": "Mesh"},
			"rules": [{"default": {"allow": [{"spiffeID": {"type": "Exact", "value": "spiffe://trust-domain/ns/default"}}]}}]
		}
	}`

	It("should accept a resource with only schema fields", func() {
		// when
		res, err := rest.JSON.UnmarshalStrict([]byte(valid), desc)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(res.GetSpec()).ToNot(BeNil())
	})

	It("should accept server-side meta fields", func() {
		// when
		res, err := rest.JSON.UnmarshalStrict([]byte(`{
			"type": "MeshTrafficPermission",
			"name": "mtp-1",
			"mesh": "default",
			"kri": "kri_mtp_default__mtp-1_",
			"creationTime": "2018-07-17T16:05:36.995Z",
			"modificationTime": "2018-07-17T16:05:36.995Z",
			"spec": {
				"targetRef": {"kind": "Mesh"},
				"rules": [{"default": {"allow": [{"spiffeID": {"type": "Exact", "value": "spiffe://trust-domain/ns/default"}}]}}]
			}
		}`), desc)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(res).ToNot(BeNil())
	})

	It("should reject a field that was removed from the schema", func() {
		// given a document with spec.from that was replaced by spec.rules
		doc := `{
			"type": "MeshTrafficPermission",
			"name": "mtp-1",
			"mesh": "default",
			"spec": {
				"targetRef": {"kind": "Mesh"},
				"from": [{"targetRef": {"kind": "Mesh"}, "default": {"action": "Allow"}}],
				"rules": [{"default": {"allow": [{"spiffeID": {"type": "Exact", "value": "spiffe://trust-domain/ns/default"}}]}}]
			}
		}`

		// when
		_, err := rest.JSON.UnmarshalStrict([]byte(doc), desc)

		// then
		Expect(err).To(HaveOccurred())
		Expect(validators.IsValidationError(err)).To(BeTrue())
		Expect(err.(*validators.ValidationError).Violations).To(Equal([]validators.Violation{
			{Field: "spec.from", Message: "unknown field"},
		}))
	})

	It("should reject unknown fields at any level", func() {
		// given
		doc := `{
			"type": "MeshTrafficPermission",
			"name": "mtp-1",
			"mesh": "default",
			"bogus": true,
			"spec": {
				"targetRef": {"kind": "Mesh"},
				"rules": [{"default": {"allow": [{"spiffeID": {"type": "Exact", "value": "spiffe://trust-domain/ns/default"}, "bogus": 1}]}}]
			}
		}`

		// when
		_, err := rest.JSON.UnmarshalStrict([]byte(doc), desc)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.(*validators.ValidationError).Violations).To(Equal([]validators.Violation{
			{Field: "bogus", Message: "unknown field"},
			{Field: "spec.rules[0].default.allow[0].bogus", Message: "unknown field"},
		}))
	})

	It("should reject fields from the Kubernetes resource representation", func() {
		// given a document with the k8s-style wrapper that the REST representation inlines
		doc := `{
			"apiVersion": "kuma.io/v1alpha1",
			"kind": "MeshTrafficPermission",
			"metadata": {"labels": {"team": "a"}},
			"type": "MeshTrafficPermission",
			"name": "mtp-1",
			"mesh": "default",
			"spec": {
				"targetRef": {"kind": "Mesh"},
				"rules": [{"default": {"allow": [{"spiffeID": {"type": "Exact", "value": "spiffe://trust-domain/ns/default"}}]}}]
			}
		}`

		// when
		_, err := rest.JSON.UnmarshalStrict([]byte(doc), desc)

		// then
		Expect(err).To(HaveOccurred())
		Expect(validators.IsValidationError(err)).To(BeTrue())
		Expect(err.(*validators.ValidationError).Violations).To(Equal([]validators.Violation{
			{Field: "apiVersion", Message: "unknown field"},
			{Field: "kind", Message: "unknown field"},
			{Field: "metadata", Message: "unknown field"},
		}))
	})

	It("should keep ignoring unknown fields on lenient unmarshal", func() {
		// given
		doc := `{
			"type": "MeshTrafficPermission",
			"name": "mtp-1",
			"mesh": "default",
			"spec": {
				"targetRef": {"kind": "Mesh"},
				"from": [{"targetRef": {"kind": "Mesh"}, "default": {"action": "Allow"}}],
				"rules": [{"default": {"allow": [{"spiffeID": {"type": "Exact", "value": "spiffe://trust-domain/ns/default"}}]}}]
			}
		}`

		// when
		res, err := rest.JSON.Unmarshal([]byte(doc), desc)

		// then
		Expect(err).ToNot(HaveOccurred())
		spec := res.GetSpec().(*mtp_api.MeshTrafficPermission)
		Expect(spec.Rules).ToNot(BeNil())
		Expect(*spec.Rules).To(HaveLen(1))
	})

	It("should reject unknown fields in YAML documents", func() {
		// given
		doc := `
type: MeshTrafficPermission
name: mtp-1
mesh: default
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        action: Allow
  rules:
    - default:
        allow:
          - spiffeID:
              type: Exact
              value: spiffe://trust-domain/ns/default
`

		// when
		_, err := rest.YAML.UnmarshalCoreStrict([]byte(doc))

		// then
		Expect(err).To(HaveOccurred())
		Expect(validators.IsValidationError(err)).To(BeTrue())
		Expect(err.(*validators.ValidationError).Violations).To(Equal([]validators.Violation{
			{Field: "spec.from", Message: "unknown field"},
		}))
	})

	It("should not flag content of free-form schema fields", func() {
		// when
		res, err := rest.YAML.UnmarshalCoreStrict([]byte(`
type: MeshAccessLog
name: mal-1
mesh: default
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: Mesh
      default:
        backends:
          - type: OpenTelemetry
            openTelemetry:
              backendRef:
                kind: MeshOpenTelemetryBackend
                labels:
                  app: otel-collector
              attributes:
                - key: mesh
                  value: "%KUMA_MESH%"
              body:
                kvlistValue:
                  values:
                    - key: mesh
                      value:
                        stringValue: "%KUMA_MESH%"
`))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Descriptor().Name).To(Equal(core_model.ResourceType("MeshAccessLog")))
	})

	It("should accept a valid YAML document with strict unmarshal", func() {
		// when
		res, err := rest.YAML.UnmarshalCoreStrict([]byte(`
type: MeshTrafficPermission
name: mtp-1
mesh: default
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        allow:
          - spiffeID:
              type: Exact
              value: spiffe://trust-domain/ns/default
`))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Descriptor().Name).To(Equal(core_model.ResourceType("MeshTrafficPermission")))
	})
})

// The spec describes unions as a oneOf of named members, which the structural
// schema validating requests cannot follow. These guard that flattening them
// back keeps validating the union's properties.
var _ = Describe("UnmarshalStrict of a discriminated union", func() {
	desc := mlbs_api.MeshLoadBalancingStrategyResourceTypeDescriptor

	withLoadBalancer := func(loadBalancer string) []byte {
		return []byte(`{
			"type": "MeshLoadBalancingStrategy",
			"name": "mlbs-1",
			"mesh": "default",
			"spec": {
				"targetRef": {"kind": "Mesh"},
				"to": [{"targetRef": {"kind": "Mesh"}, "default": {"loadBalancer": ` + loadBalancer + `}}]
			}
		}`)
	}

	DescribeTable("should accept",
		func(loadBalancer string) {
			_, err := rest.JSON.UnmarshalStrict(withLoadBalancer(loadBalancer), desc)

			Expect(err).ToNot(HaveOccurred())
		},
		Entry("a variant with no config", `{"type": "RoundRobin"}`),
		Entry("a variant with its config", `{"type": "RingHash", "ringHash": {"minRingSize": 1024}}`),
	)

	It("should reject an unknown field inside a variant", func() {
		_, err := rest.JSON.UnmarshalStrict(withLoadBalancer(`{"type": "RingHash", "ringHash": {"bogus": 1}}`), desc)

		Expect(err).To(HaveOccurred())
		Expect(err.(*validators.ValidationError).Violations).To(Equal([]validators.Violation{
			{Field: "spec.to[0].default.loadBalancer.ringHash.bogus", Message: "unknown field"},
		}))
	})

	It("should reject a value the discriminator does not allow", func() {
		_, err := rest.JSON.UnmarshalStrict(withLoadBalancer(`{"type": "Bogus"}`), desc)

		Expect(err).To(HaveOccurred())
		Expect(validators.IsValidationError(err)).To(BeTrue())
	})
})
