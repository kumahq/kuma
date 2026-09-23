package context_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/metadata"

	"github.com/kumahq/kuma/v3/pkg/config/multizone"
	kds_context "github.com/kumahq/kuma/v3/pkg/kds/context"
)

type noopAuthenticator struct{}

func (noopAuthenticator) Authenticate(context.Context, metadata.MD, string) error {
	return nil
}

var _ = Describe("RegisterZoneAuthenticator", func() {
	var kdsContext *kds_context.Context

	BeforeEach(func() {
		kdsContext = &kds_context.Context{}
	})

	It("should register an authenticator of a distribution next to the Kuma ones", func() {
		Expect(kdsContext.RegisterZoneAuthenticator(multizone.KDSAuthZoneToken, noopAuthenticator{})).To(Succeed())
		Expect(kdsContext.RegisterZoneAuthenticator("custom", noopAuthenticator{})).To(Succeed())

		Expect(kdsContext.ZoneAuthenticators).To(HaveLen(2))
	})

	It("should refuse to replace the authenticator of a registered type", func() {
		Expect(kdsContext.RegisterZoneAuthenticator(multizone.KDSAuthZoneToken, noopAuthenticator{})).To(Succeed())

		err := kdsContext.RegisterZoneAuthenticator(multizone.KDSAuthZoneToken, noopAuthenticator{})

		Expect(err).To(MatchError(`authenticator for KDS auth type "zoneToken" is already registered`))
	})

	It("should refuse to authenticate the none type", func() {
		err := kdsContext.RegisterZoneAuthenticator(multizone.KDSAuthNone, noopAuthenticator{})

		Expect(err).To(MatchError(ContainSubstring("it disables authentication of Zone CPs")))
		Expect(kdsContext.ZoneAuthenticators).To(BeEmpty())
	})
})
