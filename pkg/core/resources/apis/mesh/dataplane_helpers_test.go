package mesh_test

import (
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/core/resources/apis/mesh"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("Dataplane", func() {

	Describe("UsesInterface()", func() {

		type testCase struct {
			dataplane string
			address   string
			port      uint32
			expected  bool
		}

		DescribeTable("should correctly determine whether a given (ip, port) endpoint would overshadow one of Dataplane interfaces",
			func(given testCase) {
				// given
				dataplane := &DataplaneResource{}

				// when
				Expect(util_proto.FromYAML([]byte(given.dataplane), &dataplane.Spec)).To(Succeed())
				// then
				Expect(dataplane.UsesInterface(net.ParseIP(given.address), given.port)).To(Equal(given.expected))
			},
			Entry("port of an inbound interface is overshadowed (wilcard ip match)", testCase{
				dataplane: `
                networking:
                  inbound:
                  - interface: 192.168.0.1:80:8080
                  outbound:
                  - interface: :54321
                    service: db
                  - interface: :59200
                    service: elastic
`,
				address:  "0.0.0.0",
				port:     80,
				expected: true,
			}),
			Entry("port of the application is overshadowed (wilcard ip match)", testCase{
				dataplane: `
                networking:
                  inbound:
                   - interface: 192.168.0.1:80:8080
                  outbound:
                  - interface: :54321
                    service: db
                  - interface: :59200
                    service: elastic
`,
				address:  "0.0.0.0",
				port:     8080,
				expected: true,
			}),
			Entry("port of an outbound interface is overshadowed (wilcard ip match)", testCase{
				dataplane: `
                networking:
                  inbound:
                   - interface: 192.168.0.1:80:8080
                  outbound:
                  - interface: :54321
                    service: db
                  - interface: :59200
                    service: elastic
`,
				address:  "0.0.0.0",
				port:     54321,
				expected: true,
			}),
			Entry("port of an inbound interface is overshadowed (exact ip match)", testCase{
				dataplane: `
                networking:
                  inbound:
                  - interface: 192.168.0.1:80:8080
                  outbound:
                  - interface: :54321
                    service: db
                  - interface: :59200
                    service: elastic
`,
				address:  "192.168.0.1",
				port:     80,
				expected: true,
			}),
			Entry("port of the application is overshadowed (exact ip match)", testCase{
				dataplane: `
                networking:
                  inbound:
                   - interface: 192.168.0.1:80:8080
                  outbound:
                  - interface: :54321
                    service: db
                  - interface: :59200
                    service: elastic
`,
				address:  "127.0.0.1",
				port:     8080,
				expected: true,
			}),
			Entry("port of an outbound interface is overshadowed (exact ip match)", testCase{
				dataplane: `
                networking:
                  inbound:
                   - interface: 192.168.0.1:80:8080
                  outbound:
                  - interface: :54321
                    service: db
                  - interface: :59200
                    service: elastic
`,
				address:  "127.0.0.1",
				port:     54321,
				expected: true,
			}),
			Entry("port of invalid inbound interface is not overshadowed", testCase{
				dataplane: `
                networking:
                  inbound:
                   - interface: ?:80:8080
                  outbound:
                  - interface: :54321
                    service: db
                  - interface: :59200
                    service: elastic
`,
				address:  "0.0.0.0",
				port:     80,
				expected: false,
			}),
			Entry("port of invalid outbound interface is not overshadowed", testCase{
				dataplane: `
                networking:
                  inbound:
                   - interface: 192.168.0.1:80:8080
                  outbound:
                  - interface: ?:54321
                    service: db
                  - interface: :59200
                    service: elastic
`,
				address:  "0.0.0.0",
				port:     54321,
				expected: false,
			}),
			Entry("non-overlapping ports are not overshadowed", testCase{
				dataplane: `
                networking:
                  inbound:
                   - interface: 192.168.0.1:80:8080
                  outbound:
                  - interface: :54321
                    service: db
                  - interface: :59200
                    service: elastic
`,
				address:  "0.0.0.0",
				port:     5670,
				expected: false,
			}),
			Entry("non-overlapping ip addresses are not overshadowed (inbound listener port)", testCase{
				dataplane: `
                networking:
                  inbound:
                   - interface: 192.168.0.1:80:8080
                  outbound:
                  - interface: :54321
                    service: db
                  - interface: :59200
                    service: elastic
`,
				address:  "192.168.0.2",
				port:     80,
				expected: false,
			}),
			Entry("non-overlapping ip addresses are not overshadowed (application port)", testCase{
				dataplane: `
                networking:
                  inbound:
                   - interface: 192.168.0.1:80:8080
                  outbound:
                  - interface: :54321
                    service: db
                  - interface: :59200
                    service: elastic
`,
				address:  "192.168.0.2",
				port:     8080,
				expected: false,
			}),
			Entry("non-overlapping ip addresses are not overshadowed (outbound listener port)", testCase{
				dataplane: `
                networking:
                  inbound:
                   - interface: 192.168.0.1:80:8080
                  outbound:
                  - interface: :54321
                    service: db
                  - interface: :59200
                    service: elastic
`,
				address:  "192.168.0.2",
				port:     54321,
				expected: false,
			}),
		)

		It("should not crash if Dataplane is nil", func() {
			// given
			var dataplane *DataplaneResource
			address := net.ParseIP("0.0.0.0")
			port := uint32(5670)
			expected := false
			// expect
			Expect(dataplane.UsesInterface(address, port)).To(Equal(expected))
		})
	})
})
