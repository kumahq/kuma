package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	. "github.com/kumahq/kuma/v3/pkg/test/resources/validators"
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
		ErrorCase("spec.extension.type empty",
			validators.Violation{
				Field:   `spec.extension.type`,
				Message: `must not be empty`,
			}, `
template: "{{ .Name }}.mesh"
selector:
  meshService: {}
extension:
  config: {}
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
		ErrorCase("spec.template renders to uppercase",
			validators.Violation{
				Field:   `spec.template`,
				Message: `template renders to "name.MESH" which is not a valid DNS name: a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`,
			}, `
template: "{{ .Name }}.MESH"
selector:
  meshService: {}
`),
		ErrorCase("spec.template renders with leading dot",
			validators.Violation{
				Field:   `spec.template`,
				Message: `template renders to ".name.mesh" which is not a valid DNS name: a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`,
			}, `
template: ".{{ .Name }}.mesh"
selector:
  meshService: {}
`),
		ErrorCase("spec.template renders with consecutive dots",
			validators.Violation{
				Field:   `spec.template`,
				Message: `template renders to "name..mesh" which is not a valid DNS name: a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`,
			}, `
template: "{{ .Name }}..mesh"
selector:
  meshService: {}
`),
		ErrorCase("spec.template renders with underscore",
			validators.Violation{
				Field:   `spec.template`,
				Message: `template renders to "name_svc.mesh" which is not a valid DNS name: a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`,
			}, `
template: "{{ .Name }}_svc.mesh"
selector:
  meshService: {}
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
		Entry("accepts template using all builtin fields", `
template: "{{ .Name }}.{{ .DisplayName }}.{{ .Namespace }}.{{ .Mesh }}.{{ .Zone }}.mesh"
selector:
  meshService: {}
`),
		Entry("accepts template using label function", `
template: "{{ label \"app\" }}.mesh"
selector:
  meshService: {}
`),
	)
})
