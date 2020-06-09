package ingress_test

import (
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	"github.com/Kong/kuma/pkg/xds/ingress"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ingress Dataplane", func() {

	type testCase struct {
		dataplanes []string
		expected   string
	}
	DescribeTable("should generate ingress based on other dataplanes",
		func(given testCase) {
			dataplanes := []*core_mesh.DataplaneResource{}

			for _, dp := range given.dataplanes {
				dpRes := &core_mesh.DataplaneResource{}
				err := util_proto.FromYAML([]byte(dp), &dpRes.Spec)
				Expect(err).ToNot(HaveOccurred())
				dataplanes = append(dataplanes, dpRes)
			}

			actual := ingress.GetIngressByDataplanes(dataplanes)
			actualYAML, err := yaml.Marshal(actual)
			Expect(err).ToNot(HaveOccurred())
			Expect(actualYAML).To(MatchYAML(given.expected))
		},
		Entry("base", testCase{
			dataplanes: []string{
				`
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              inbound:
                - address: 127.0.0.1
                  port: 1010
                  servicePort: 2020
                  tags:
                    service: backend
                    version: "1"
                    region: eu
`,
				`
            type: Dataplane
            name: dp-2
            mesh: default
            networking:
              inbound:
                - address: 127.0.0.1
                  port: 1010
                  servicePort: 2020
                  tags:
                    service: backend
                    version: "2"
                    region: us
`,
			},
			expected: `
            - service: backend
              tags:
                region: eu
                version: "1"
            - service: backend
              tags:
                region: us
                version: "2"
`,
		}))
	Entry("duplicate service tags", testCase{
		dataplanes: []string{
			`
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              inbound:
                - address: 127.0.0.1
                  port: 1010
                  servicePort: 2020
                  tags:
                    service: backend
`,
			`
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              inbound:
                - address: 1.1.1.1
                  port: 1010
                  servicePort: 2020
                  tags:
                    service: backend
`,
			`
            type: Dataplane
            name: dp-2
            mesh: default
            networking:
              inbound:
                - address: 127.0.0.1
                  port: 1010
                  servicePort: 2020
                  tags:
                    service: backend
                    version: "2"
                    region: us
`,
		},
		expected: `
            - service: backend
            - service: backend
              tags:
                region: us
                version: "2"
`,
	})
})
