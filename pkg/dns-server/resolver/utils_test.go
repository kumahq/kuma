package resolver

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DNS utils", func() {
	resolver, _ := newSimpleDNSResolver("127.0.0.1", "5653", "240.0.0.0/4")

	It("domainFromName", func(done Done) {
		_, error := resolver.domainFromName("")
		Expect(error).To(HaveOccurred())

		_, error = resolver.domainFromName("kuma")
		Expect(error).ToNot(HaveOccurred())

		domain, error := resolver.domainFromName(".kuma")
		Expect(error).ToNot(HaveOccurred())
		Expect(domain).To(Equal("kuma"))

		domain, error = resolver.domainFromName("namespace.kuma")
		Expect(error).ToNot(HaveOccurred())
		Expect(domain).To(Equal("kuma"))

		domain, error = resolver.domainFromName("service.namespace.kuma")
		Expect(error).ToNot(HaveOccurred())
		Expect(domain).To(Equal("kuma"))

		domain, error = resolver.domainFromName("tag.service.namespace.kuma")
		Expect(error).ToNot(HaveOccurred())
		Expect(domain).To(Equal("kuma"))

		// ready
		close(done)
	})

	It("serviceFromName", func(done Done) {
		_, error := resolver.serviceFromName("")
		Expect(error).To(HaveOccurred())

		_, error = resolver.serviceFromName("kuma")
		Expect(error).To(HaveOccurred())

		_, error = resolver.serviceFromName(".kuma")
		Expect(error).To(HaveOccurred())

		service, error := resolver.serviceFromName("namespace.kuma")
		Expect(error).ToNot(HaveOccurred())
		Expect(service).To(Equal("namespace"))

		service, error = resolver.serviceFromName("service.namespace.kuma")
		Expect(error).ToNot(HaveOccurred())
		Expect(service).To(Equal("service.namespace"))

		service, error = resolver.serviceFromName("tag.service.namespace.kuma")
		Expect(error).ToNot(HaveOccurred())
		Expect(service).To(Equal("tag.service.namespace"))

		// ready
		close(done)
	})
})
