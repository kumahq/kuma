package vips_test

import (
	"fmt"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/dns/vips"
)

var _ = Describe("global view", func() {

	It("should fail when no more ips", func() {
		// given
		gv, err := vips.NewGlobalView("192.168.0.1/32")
		Expect(err).ToNot(HaveOccurred())
		Expect(gv).ToNot(BeNil())

		// when
		ip1, err := gv.Allocate(vips.NewServiceEntry("foo.bar"))
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(ip1).ToNot(Equal(""))

		// when
		ip2, err := gv.Allocate(vips.NewServiceEntry("bar.bar"))
		// then
		Expect(err).To(HaveOccurred())
		Expect(ip2).To(Equal(""))
	})

	It("should allocate IPs", func() {
		// given
		gv, err := vips.NewGlobalView("192.168.0.1/24")
		Expect(err).ToNot(HaveOccurred())
		Expect(gv).ToNot(BeNil())

		// when
		ip1, err := gv.Allocate(vips.NewServiceEntry("foo.bar"))
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		ip2, err := gv.Allocate(vips.NewServiceEntry("bar.bar"))
		// then
		Expect(err).ToNot(HaveOccurred())

		// then
		vips := map[vips.HostnameEntry]string{
			vips.NewServiceEntry("foo.bar"): ip1,
			vips.NewServiceEntry("bar.bar"): ip2,
		}

		Expect(gv.ToVIPMap()).To(Equal(vips))
	})

	It("should allocate 2^16 IP addresses", func() {
		// given
		gv, err := vips.NewGlobalView("240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())
		Expect(gv).ToNot(BeNil())

		for i := 0; i < math.MaxInt16; i++ {
			// when
			_, err := gv.Allocate(vips.NewHostEntry(fmt.Sprintf("foo-%d.mesh", i)))
			// then
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("should give the same ip if we ask for a host already allocated", func() {
		// given
		host := vips.NewServiceEntry("foo.com")
		gv, err := vips.NewGlobalView("240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())
		Expect(gv).ToNot(BeNil())

		err = gv.Reserve(host, "240.0.0.1")
		Expect(err).ToNot(HaveOccurred())

		// when
		ip, err := gv.Allocate(host)
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(ip).To(Equal("240.0.0.1"))
	})

	It("should give the same ip if when allocating it twice", func() {
		// given
		host := vips.NewServiceEntry("foo.com")
		gv, err := vips.NewGlobalView("240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())
		Expect(gv).ToNot(BeNil())

		ip, err := gv.Allocate(host)
		Expect(err).ToNot(HaveOccurred())

		// when
		ip2, err := gv.Allocate(host)
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(ip2).To(Equal(ip))
	})
})
