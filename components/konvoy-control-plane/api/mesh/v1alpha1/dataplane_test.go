package v1alpha1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"

	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/internal/util/proto"
)

var _ = Describe("Dataplane", func() {

	It("should be possible to unmarshal from YAML", func() {
		// given
		input := `
        networking:
          inbound:
          - interface: 1.1.1.1:80:8080
            tags:
              service: mobile
              version: "0.1"
              env: production
          outbound:
          - interface: :30000
            service: postgres
            servicePort: 5432
          - interface: :50000
            service: redis.default.svc
            servicePort: 8000
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
		Expect(dataplane.Networking.Inbound).To(HaveLen(1))
		Expect(dataplane.Networking.Inbound[0].Interface).To(Equal("1.1.1.1:80:8080"))
		Expect(dataplane.Networking.Inbound[0].Tags).To(HaveLen(3))
		Expect(dataplane.Networking.Inbound[0].Tags).To(HaveKeyWithValue("service", "mobile"))
		Expect(dataplane.Networking.Inbound[0].Tags).To(HaveKeyWithValue("version", "0.1"))
		Expect(dataplane.Networking.Inbound[0].Tags).To(HaveKeyWithValue("env", "production"))
		// and
		Expect(dataplane.Networking.Outbound).To(HaveLen(2))
		Expect(dataplane.Networking.Outbound[0].Interface).To(Equal(":30000"))
		Expect(dataplane.Networking.Outbound[0].Service).To(Equal("postgres"))
		Expect(dataplane.Networking.Outbound[0].ServicePort).To(Equal(uint32(5432)))
		Expect(dataplane.Networking.Outbound[1].Interface).To(Equal(":50000"))
		Expect(dataplane.Networking.Outbound[1].Service).To(Equal("redis.default.svc"))
		Expect(dataplane.Networking.Outbound[1].ServicePort).To(Equal(uint32(8000)))
	})
})
