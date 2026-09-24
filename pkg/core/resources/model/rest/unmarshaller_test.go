package rest_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
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
