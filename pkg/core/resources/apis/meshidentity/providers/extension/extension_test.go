package extension_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshidentity_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/providers"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/providers/extension"
	"github.com/kumahq/kuma/v2/pkg/core/xds"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
)

type stubHandler struct {
	validateErr          error
	meshTrustCA          []byte
	validateCalled       bool
	initCalled           bool
	createCalled         bool
	getMeshTrustCACalled bool
}

var _ providers.IdentityProvider = &stubHandler{}

func (s *stubHandler) Validate(_ context.Context, _ *meshidentity_api.MeshIdentityResource) error {
	s.validateCalled = true
	return s.validateErr
}

func (s *stubHandler) Initialize(_ context.Context, _ *meshidentity_api.MeshIdentityResource) error {
	s.initCalled = true
	return nil
}

func (s *stubHandler) CreateIdentity(_ context.Context, _ *meshidentity_api.MeshIdentityResource, _ *xds.Proxy) (*xds.WorkloadIdentity, error) {
	s.createCalled = true
	return nil, nil
}

func (s *stubHandler) GetMeshTrustCA(_ context.Context, _ *meshidentity_api.MeshIdentityResource) ([]byte, error) {
	s.getMeshTrustCACalled = true
	return s.meshTrustCA, nil
}

func identityWithExtension(name string) *meshidentity_api.MeshIdentityResource {
	mid := builders.MeshIdentity().Build()
	mid.Spec.Provider = &meshidentity_api.Provider{
		Type: meshidentity_api.ExtensionType,
		Extension: &meshidentity_api.Extension{
			Name: name,
		},
	}
	return mid
}

var _ = Describe("Extension Dispatcher", func() {
	It("routes to correct handler", func() {
		acm := &stubHandler{meshTrustCA: []byte("acm-ca")}
		cm := &stubHandler{meshTrustCA: []byte("cm-ca")}
		d := extension.NewDispatcher(map[string]providers.IdentityProvider{
			"acmpca":      acm,
			"certmanager": cm,
		})

		ca, err := d.GetMeshTrustCA(context.Background(), identityWithExtension("acmpca"))

		Expect(err).ToNot(HaveOccurred())
		Expect(ca).To(Equal([]byte("acm-ca")))
		Expect(acm.getMeshTrustCACalled).To(BeTrue())
		Expect(cm.getMeshTrustCACalled).To(BeFalse())
	})

	It("fails on unknown extension name", func() {
		d := extension.NewDispatcher(map[string]providers.IdentityProvider{
			"acmpca": &stubHandler{},
		})

		err := d.Validate(context.Background(), identityWithExtension("unknown"))

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unknown extension \"unknown\""))
		Expect(err.Error()).To(ContainSubstring("acmpca"))
	})

	It("fails on nil extension", func() {
		d := extension.NewDispatcher(map[string]providers.IdentityProvider{})
		mid := builders.MeshIdentity().Build()
		mid.Spec.Provider = &meshidentity_api.Provider{
			Type: meshidentity_api.ExtensionType,
		}

		err := d.Validate(context.Background(), mid)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("extension config is nil"))
	})

	It("propagates handler error", func() {
		h := &stubHandler{validateErr: errors.New("bad config")}
		d := extension.NewDispatcher(map[string]providers.IdentityProvider{
			"acmpca": h,
		})

		err := d.Validate(context.Background(), identityWithExtension("acmpca"))

		Expect(err).To(MatchError("bad config"))
		Expect(h.validateCalled).To(BeTrue())
	})

	It("calls all methods correctly", func() {
		h := &stubHandler{meshTrustCA: []byte("ca")}
		d := extension.NewDispatcher(map[string]providers.IdentityProvider{
			"test": h,
		})
		mid := identityWithExtension("test")
		ctx := context.Background()

		Expect(d.Validate(ctx, mid)).To(Succeed())
		Expect(h.validateCalled).To(BeTrue())

		Expect(d.Initialize(ctx, mid)).To(Succeed())
		Expect(h.initCalled).To(BeTrue())

		_, err := d.CreateIdentity(ctx, mid, &xds.Proxy{})
		Expect(err).ToNot(HaveOccurred())
		Expect(h.createCalled).To(BeTrue())

		ca, err := d.GetMeshTrustCA(ctx, mid)
		Expect(err).ToNot(HaveOccurred())
		Expect(ca).To(Equal([]byte("ca")))
		Expect(h.getMeshTrustCACalled).To(BeTrue())
	})
})
