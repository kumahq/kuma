package tls_test

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	intercp_tls "github.com/kumahq/kuma/pkg/intercp/tls"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("TLS", func() {
	var ch chan struct{}
	var defaults intercp_tls.DefaultsComponent
	var resManager manager.ResourceManager

	BeforeEach(func() {
		resManager = manager.NewResourceManager(memory.NewStore())
		defaults = intercp_tls.DefaultsComponent{
			ResManager: resManager,
			Log:        logr.Discard(),
		}
		ch = make(chan struct{})
	})

	AfterEach(func() {
		close(ch)
	})

	It("should generate default CA", func() {
		// when
		Expect(defaults.Start(ch)).To(Succeed())

		// then
		_, err := intercp_tls.LoadCA(context.Background(), resManager)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should ignore creating CA when there is one already in place", func() {
		// given
		Expect(defaults.Start(ch)).To(Succeed())
		ca1, err := intercp_tls.LoadCA(context.Background(), resManager)
		Expect(err).ToNot(HaveOccurred())

		// when
		Expect(defaults.Start(ch)).To(Succeed())

		// then
		ca2, err := intercp_tls.LoadCA(context.Background(), resManager)
		Expect(err).ToNot(HaveOccurred())
		Expect(ca1).To(Equal(ca2))
	})
})
