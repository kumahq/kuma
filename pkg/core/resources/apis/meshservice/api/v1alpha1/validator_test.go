package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	. "github.com/kumahq/kuma/v3/pkg/test/resources/validators"
)

var _ = Describe("MeshService", func() {
	DescribeErrorCases(
		api.NewMeshServiceResource,
		Entry(
			"name too long",
			ResourceValidationCase{
				Violations: []validators.Violation{{
					Field:   `name`,
					Message: `must not be longer than 63 characters`,
				}},
				Name:     "meshservice-too-long-too-long-too-long-too-long-too-long-too-long-too-long-too-long-too-long",
				Resource: "",
			},
		),
		Entry(
			"multiple selectors specified",
			ResourceValidationCase{
				Violations: []validators.Violation{{
					Field:   `spec.selector`,
					Message: `must specify only one of: dataplaneRef or dataplaneLabels`,
				}},
				Name: "meshservice",
				Resource: `
selector:
  dataplaneRef:
    name: redis-01
  dataplaneLabels:
    matchLabels:
      app: redis
`,
			},
		),
	)
	DescribeValidCases(
		api.NewMeshServiceResource,
		Entry(
			"accepts valid resource",
			ResourceValidationCase{
				Name:     "meshservice",
				Resource: "",
			},
		),
		Entry(
			"accepts dataplaneRef selector",
			ResourceValidationCase{
				Name: "meshservice",
				Resource: `
selector:
  dataplaneRef:
    name: redis-01
`,
			},
		),
		Entry(
			"accepts dataplaneLabels selector",
			ResourceValidationCase{
				Name: "meshservice",
				Resource: `
selector:
  dataplaneLabels:
    matchLabels:
      app: redis
`,
			},
		),
	)

	Describe("Deprecations()", func() {
		newMeshService := func(spec string) *api.MeshServiceResource {
			ms := api.NewMeshServiceResource()
			ms.SetMeta(&test_model.ResourceMeta{Mesh: "default", Name: "redis"})
			Expect(core_model.FromYAML([]byte(spec), ms.Spec)).To(Succeed())
			return ms
		}

		It("warns when no selector is set", func() {
			ms := newMeshService(`
ports:
  - port: 6379
    targetPort: 6379
`)

			Expect(ms.Deprecations()).To(ContainElement(ContainSubstring("has no selector")))
		})

		It("does not warn when dataplaneLabels is set", func() {
			ms := newMeshService(`
selector:
  dataplaneLabels:
    matchLabels:
      app: redis
ports:
  - port: 6379
    targetPort: 6379
`)

			Expect(ms.Deprecations()).To(BeEmpty())
		})

		It("does not warn when dataplaneRef is set", func() {
			ms := newMeshService(`
selector:
  dataplaneRef:
    name: redis-01
ports:
  - port: 6379
    targetPort: 6379
`)

			Expect(ms.Deprecations()).To(BeEmpty())
		})
	})
})
