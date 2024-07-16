package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	. "github.com/kumahq/kuma/pkg/test/resources/validators"
)

var _ = Describe("validation", func() {
	DescribeErrorCases(
		api.NewHostnameGeneratorResource,
		ErrorCase("spec.template error",
			validators.Violation{
				Field:   `spec.template`,
				Message: `couldn't parse template: template: :1: bad character U+005B '['`,
			}, `
template: "{{ .Name[4 }}.mesh"
selector:
  meshService: {}
`),
		ErrorCase("spec.template empty",
			validators.Violation{
				Field:   `spec.template`,
				Message: `must not be empty`,
			}, `
template: ""
selector:
  meshService: {}
`),
		ErrorCase("spec.selector empty",
			validators.Violation{
				Field:   `spec.selector`,
				Message: `exact one selector (meshService, meshExternalService) must be defined`,
			}, `
template: "{{ .Name }}.mesh"
`),
		ErrorCase("spec.selector has too many selectors",
			validators.Violation{
				Field:   `spec.selector`,
				Message: `exact one selector (meshService, meshExternalService) must be defined`,
			}, `
template: "{{ .Name }}.mesh"
selector:
  meshService: {}
  meshExternalService: {}
`),
	)
	DescribeValidCases(
		api.NewHostnameGeneratorResource,
		Entry("accepts valid resource", `
template: "{{ .Name }}.mesh"
selector:
  meshService: {}
`),
		Entry("accepts valid resource", `
template: "{{ .Name }}.mesh"
selector:
  meshMultiZoneService: {}
`),
	)
})
