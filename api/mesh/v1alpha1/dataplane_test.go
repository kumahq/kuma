package v1alpha1_test

import (
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

var _ = Describe("Dataplane", func() {
	It("should be possible to unmarshal from YAML", func() {
		// given
		input := `
        networking:
          address: 1.1.1.1
          inbound:
          - port: 80
            servicePort: 8080
            address: 2.2.2.2
            protocol: http
          outbound:
          - port: 30000
            backendRef:
              kind: MeshService
              name: postgres
              port: 5432
          - port: 50000
            backendRef:
              kind: MeshService
              name: redis
              port: 6379
`
		// when
		dataplane := &Dataplane{}
		err := util_proto.FromYAML([]byte(input), dataplane)
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(dataplane.Networking.Address).To(Equal("1.1.1.1"))
		Expect(dataplane.Networking.Inbound).To(HaveLen(1))
		Expect(dataplane.Networking.Inbound[0].Port).To(Equal(uint32(80)))
		Expect(dataplane.Networking.Inbound[0].ServicePort).To(Equal(uint32(8080)))
		Expect(dataplane.Networking.Inbound[0].Address).To(Equal("2.2.2.2"))
		Expect(dataplane.Networking.Inbound[0].Protocol).To(Equal("http"))
		Expect(dataplane.Networking.Outbound).To(HaveLen(2))
		Expect(dataplane.Networking.Outbound[0].Port).To(Equal(uint32(30000)))
		Expect(dataplane.Networking.Outbound[0].BackendRef.Name).To(Equal("postgres"))
		Expect(dataplane.Networking.Outbound[1].Port).To(Equal(uint32(50000)))
		Expect(dataplane.Networking.Outbound[1].BackendRef.Name).To(Equal("redis"))
	})

	Describe("json.Marshal()", func() {
		type testCase struct {
			input    string
			expected string
		}

		DescribeTable("should serialize fields in the correct order",
			func(given testCase) {
				// given
				dataplane := &Dataplane{}

				// when
				err := util_proto.FromYAML([]byte(given.input), dataplane)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				actual, err := util_proto.ToJSON(dataplane)
				// then
				Expect(err).ToNot(HaveOccurred())

				// given
				var pretty bytes.Buffer
				// when
				Expect(json.Indent(&pretty, actual, "", "  ")).To(Succeed())
				// and
				Expect(pretty.String()).To(Equal(given.expected))
			},
			Entry("gateway dataplane", testCase{
				input: `
                networking:
                  outbound:
                  - port: 40001
                    backendRef:
                      kind: MeshService
                      name: backend
                      port: 8080
                  inbound:
                  - port: 8080
                    protocol: http
                  address: 192.168.0.1
`,
				expected: `{
  "networking": {
    "address": "192.168.0.1",
    "inbound": [
      {
        "port": 8080,
        "protocol": "http"
      }
    ],
    "outbound": [
      {
        "port": 40001,
        "backendRef": {
          "kind": "MeshService",
          "name": "backend",
          "port": 8080
        }
      }
    ]
  }
}`,
			}),
			Entry("gateway dataplane", testCase{
				input: `
                networking:
                  outbound:
                  - port: 40001
                    backendRef:
                      kind: MeshService
                      name: backend
                      port: 8080
                  gateway:
                    tags:
                      kuma.io/service: gateway
                  address: 192.168.0.1
`,
				expected: `{
  "networking": {
    "address": "192.168.0.1",
    "gateway": {
      "tags": {
        "kuma.io/service": "gateway"
      }
    },
    "outbound": [
      {
        "port": 40001,
        "backendRef": {
          "kind": "MeshService",
          "name": "backend",
          "port": 8080
        }
      }
    ]
  }
}`,
			}),
		)
	})
})
