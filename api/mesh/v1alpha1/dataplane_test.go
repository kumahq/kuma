package v1alpha1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/api/mesh/v1alpha1"

	util_proto "github.com/Kong/kuma/api/internal/util/proto"
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
            tags:
              service: mobile
              version: "0.1"
              env: production
          outbound:
          - port: 30000
            service: postgres
          - port: 50000
            service: redis.default.svc
`
		// when
		dataplane := &Dataplane{}
		err := util_proto.FromYAML([]byte(input), dataplane)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = dataplane.Validate()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(dataplane.Networking.Address).To(Equal("1.1.1.1"))
		// and
		Expect(dataplane.Networking.Inbound).To(HaveLen(1))
		Expect(dataplane.Networking.Inbound[0].Port).To(Equal(uint32(80)))
		Expect(dataplane.Networking.Inbound[0].ServicePort).To(Equal(uint32(8080)))
		Expect(dataplane.Networking.Inbound[0].Address).To(Equal("2.2.2.2"))
		Expect(dataplane.Networking.Inbound[0].Tags).To(HaveLen(3))
		Expect(dataplane.Networking.Inbound[0].Tags).To(HaveKeyWithValue("service", "mobile"))
		Expect(dataplane.Networking.Inbound[0].Tags).To(HaveKeyWithValue("version", "0.1"))
		Expect(dataplane.Networking.Inbound[0].Tags).To(HaveKeyWithValue("env", "production"))
		// and
		Expect(dataplane.Networking.Outbound).To(HaveLen(2))
		Expect(dataplane.Networking.Outbound[0].Port).To(Equal(uint32(30000)))
		Expect(dataplane.Networking.Outbound[0].Service).To(Equal("postgres"))
		Expect(dataplane.Networking.Outbound[1].Port).To(Equal(uint32(50000)))
		Expect(dataplane.Networking.Outbound[1].Service).To(Equal("redis.default.svc"))
	})
})
