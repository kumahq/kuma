package cni

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewIntermediateConfig", func() {
	It("should set UID to default value if annotation is not specified", func() {
		a := map[string]string{}
		cfg, err := NewIntermediateConfig(a)
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.noRedirectUID).To(Equal(defaultNoRedirectUID))
	})

	It("should override UID when annotation is specified", func() {
		a := map[string]string{
			"kuma.io/sidecar-uid": "1234",
		}
		cfg, err := NewIntermediateConfig(a)
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.noRedirectUID).To(Equal("1234"))
	})
})
