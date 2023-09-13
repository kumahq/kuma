package server_test

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/kds/v2/server"
	"github.com/kumahq/kuma/pkg/kds/v2/util"
	"github.com/kumahq/kuma/pkg/multitenant"
)

var _ = Describe("Tenant callbacks", func() {
	It("should enrich metadata with tenant info", func() {
		// given
		streamID := int64(1)
		callbacks := server.NewTenancyCallbacks(&sampleTenants{})

		// when
		err := callbacks.OnDeltaStreamOpen(multitenant.WithTenant(context.Background(), "sample"), streamID, "")
		Expect(err).ToNot(HaveOccurred())

		req := &envoy_sd.DeltaDiscoveryRequest{
			Node: &envoy_core.Node{},
		}
		err = callbacks.OnStreamDeltaRequest(streamID, req)

		// then
		Expect(err).ToNot(HaveOccurred())

		tenant, ok := util.TenantFromMetadata(req.GetNode())
		Expect(ok).To(BeTrue())
		Expect(tenant).To(Equal("sample"))
	})
})

type sampleTenants struct{}

func (s sampleTenants) GetID(ctx context.Context) (string, error) {
	tenant, _ := multitenant.TenantFromCtx(ctx)
	return tenant, nil
}

func (s sampleTenants) GetIDs(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (s sampleTenants) SupportsSharding() bool {
	return false
}

func (s sampleTenants) IDSupported(ctx context.Context, id string) (bool, error) {
	return true, nil
}
