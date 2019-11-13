package mesh_test

import (
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
	"github.com/ghodss/yaml"
)

var _ = Describe("ProxyTemplate", func() {
	Describe("Validate()", func() {
		It("should pass validation", func() {
			// given
			spec := `
            selectors:
            - match:
                service: backend
            imports:
            - default-proxy
            resources:
            - name: additional
              version: v1
              resource: test`

			proxyTemplate := mesh.ProxyTemplateResource{}

			// when
			err := util_proto.FromYAML([]byte(spec), &proxyTemplate.Spec)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = proxyTemplate.Validate()
			// then
			Expect(err).ToNot(HaveOccurred())
		})

		type testCase struct {
			proxyTemplate string
			expected      string
		}
		DescribeTable("should validate fields",
			func(given testCase) {
				// given
				proxyTemplate := mesh.ProxyTemplateResource{}

				// when
				err := util_proto.FromYAML([]byte(given.proxyTemplate), &proxyTemplate.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := proxyTemplate.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty import", testCase{
				proxyTemplate: `
                imports:
                - ""
                selectors:
                - match:
                    service: backend`,
				expected: `
                violations:
                - field: imports[0]
                  message: cannot be empty`,
			}),
			Entry("unknown profile", testCase{
				proxyTemplate: `
                imports:
                - unknown-porfile
                selectors:
                - match:
                    service: backend`,
				expected: `
                violations:
                - field: imports[0]
                  message: 'profile not found. Available profiles: default-proxy'`,
			}),
			Entry("resources empty fields", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    service: backend
                resources:
                - name:
                  version:
                  resource:`,
				expected: `
                violations:
                - field: resources[0].name
                  message: cannot be empty
                - field: resources[0].version
                  message: cannot be empty
                - field: resources[0].resource
                  message: cannot be empty`,
			}),
			Entry("selector without tags", testCase{
				proxyTemplate: `
                selectors:
                - match:`,
				expected: `
                violations:
                - field: selectors[0]
                  message: has to contain at least one tag`,
			}),
			Entry("invalid tags", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    "": asdf
                    service:`,
				expected: `
                violations:
                - field: selectors[0][""]
                  message: tag cannot be empty
                - field: 'selectors[0]["service"]'
                  message: value of tag cannot be empty`,
			}),
		)
	})
})
