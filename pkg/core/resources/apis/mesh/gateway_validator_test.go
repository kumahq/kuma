package mesh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

// ResourceGenerator creates a resource of a pre-defined type.
type ResourceGenerator interface {
	New() model.Resource
}

// ResourceValidationCase captures a resource YAML and any corresponding validation error.
type ResourceValidationCase struct {
	Resource  string
	Violation validators.Violation
}

// DescribeValidCases creates a Ginkgo table test for the given entries,
// where each entry is a valid YAML resource. It ensures that each entry
// can be successfully validated.
func DescribeValidCases(generator ResourceGenerator, cases ...TableEntry) {
	DescribeTable(
		"should pass validation",
		func(given string) {
			// setup
			resource := generator.New()

			// when
			err := util_proto.FromYAML([]byte(given), resource.GetSpec())

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			verr := resource.Validate()

			// then
			Expect(verr).ToNot(HaveOccurred())
		},
		cases...)
}

// DescribeErrorCases creates a Ginkgo table test for the given entries, where each entry
// is a ResourceValidationCase that contains an invalid resource YAML and the corresponding
// validation error.
func DescribeErrorCases(generator ResourceGenerator, cases ...TableEntry) {
	DescribeTable(
		"should validate all fields and return as many individual errors as possible",
		func(given ResourceValidationCase) {
			// setup
			resource := generator.New()

			// when
			Expect(
				util_proto.FromYAML([]byte(given.Resource), resource.GetSpec()),
			).ToNot(HaveOccurred())

			expected := validators.ValidationError{
				Violations: []validators.Violation{
					given.Violation,
				}}

			// then
			Expect(resource.Validate()).To(Equal(expected.OrNil()))
		},
		cases...,
	)
}

// ErrorCase is a helper that generates a table entry for DescribeErrorCases.
func ErrorCase(description string, err validators.Violation, yaml string) TableEntry {
	return Entry(
		description,
		ResourceValidationCase{
			Violation: err,
			Resource:  yaml,
		},
	)
}

// GatewayGenerateor is a ResourceGenerator that creates GatewayResource objects.
type GatewayGenerator func() *GatewayResource

func (g GatewayGenerator) New() model.Resource {
	if g != nil {
		return g()
	}

	return nil
}

var _ = Describe("Gateway", func() {
	DescribeValidCases(
		GatewayGenerator(NewGatewayResource),
		Entry("HTTPS listener", `
type: Gateway
name: gateway
mesh: default
sources:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - hostname: www-1.example.com
    port: 443
    protocol: HTTP
    tags:
      name: https`,
		),
	)

	DescribeErrorCases(
		GatewayGenerator(NewGatewayResource),
		ErrorCase("doesn't have any source selector",
			validators.Violation{
				Field:   `sources`,
				Message: `must have at least one element`,
			}, `
type: Gateway
name: gateway
mesh: default
sources:
tags:
  product: edge
conf:
  listeners:
  - port: 443
    protocol: HTTPS
    tags:
      name: https
`),

		ErrorCase("has a service tag",
			validators.Violation{
				Field:   `tags["kuma.io/service"]`,
				Message: `tag name must not be "kuma.io/service"`,
			}, `
type: Gateway
name: gateway
mesh: default
sources:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
  kuma.io/service: gateway
conf:
  listeners:
  - port: 443
    protocol: HTTP
    tags:
      name: https
`),

		ErrorCase("doesn't have a configuration spec",
			validators.Violation{
				Field:   "conf",
				Message: "cannot be empty",
			}, `
type: Gateway
name: gateway
mesh: default
sources:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
`),

		ErrorCase("has an invalid port",
			validators.Violation{
				Field:   "conf.listeners[0].port",
				Message: "port must be in the range [1, 65535]",
			}, `
type: Gateway
name: gateway
mesh: default
sources:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - protocol: HTTP
    tags:
      name: https
`),

		ErrorCase("has an empty protocol",
			validators.Violation{
				Field:   "conf.listeners[0].protocol",
				Message: "cannot be empty",
			}, `
type: Gateway
name: gateway
mesh: default
sources:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - port: 99
    tags:
      name: https
`),

		ErrorCase("has an empty TLS mode",
			validators.Violation{
				Field:   "conf.listeners[0].tls.mode",
				Message: "cannot be empty",
			}, `
type: Gateway
name: gateway
mesh: default
sources:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - protocol: HTTPS
    port: 99
    tags:
      name: https
    tls:
      options:
`),

		ErrorCase("has a passthrough TLS secret",
			validators.Violation{
				Field:   "conf.listeners[0].tls.certificate",
				Message: "must be empty in TLS passthrough mode",
			}, `
type: Gateway
name: gateway
mesh: default
sources:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - protocol: HTTPS
    port: 99
    tags:
      name: https
    tls:
      mode: PASSTHROUGH
      certificate:
        secret: foo
`),

		ErrorCase("is missing a TLS termination secret",
			validators.Violation{
				Field:   "conf.listeners[0].tls.certificate",
				Message: "cannot be empty in TLS termination mode",
			}, `
type: Gateway
name: gateway
mesh: default
sources:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - protocol: HTTPS
    port: 99
    tags:
      name: https
    tls:
      mode: TERMINATE
`),

		ErrorCase("has an invalid wildcard",
			validators.Violation{
				Field:   "conf.listeners[0].hostname",
				Message: "invalid wildcard domain",
			}, `
type: Gateway
name: gateway
mesh: default
sources:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - hostname: "*.foo.*.example.com"
    protocol: HTTP
    port: 99
    tags:
      name: https
`),

		ErrorCase("has an invalid hostname",
			validators.Violation{
				Field:   "conf.listeners[0].hostname",
				Message: "invalid hostname",
			}, `
type: Gateway
name: gateway
mesh: default
sources:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - hostname: "foo.example$.com"
    protocol: HTTP
    port: 99
    tags:
      name: https
`),
	)
})
