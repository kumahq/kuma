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
type: HostnameGenerator
name: route-1
template: "{{ .Name[4 }}.mesh"
`),
	)
	DescribeValidCases(
		api.NewHostnameGeneratorResource,
		Entry("accepts valid resource", `
type: HostnameGenerator
name: route-1
template: "{{ .Name }}.mesh"
`),
	)
})
