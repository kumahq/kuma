package gateway_test

import (
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	xds_server "github.com/kumahq/kuma/pkg/xds/server/v3"
)

var _ = Describe("Gateway Listener", func() {
	var rt runtime.Runtime

	BeforeEach(func() {
		var err error

		rt, err = BuildRuntime()
		Expect(err).To(Succeed(), "build runtime instance")

		Expect(StoreNamedFixture(rt, "mesh-default.yaml"))
		Expect(StoreNamedFixture(rt, "dataplane-default.yaml"))
	})

	It("should not be implemented", func() {
		serverCtx := xds_server.NewXdsContext()
		reconciler := xds_server.DefaultReconciler(rt, serverCtx)

		dp, err := FetchNamedFixture(rt, core_mesh.DataplaneType,
			core_model.ResourceKey{Mesh: "default", Name: "default"})
		Expect(err).To(Succeed())

		proxy := xds.Proxy{
			Id:         xds.FromResourceKey(core_model.MetaToResourceKey(dp.GetMeta())),
			APIVersion: envoy.APIV3,
			Dataplane:  dp.(*core_mesh.DataplaneResource),
			Metadata:   nil,
			Routing:    xds.Routing{},
			Policies:   xds.MatchedPolicies{},
		}

		Expect(proxy.Dataplane.Spec.IsBuiltinGateway()).To(BeTrue())

		// The gateway listener generator isn't implemented,
		// so we expect this should fail.
		err = reconciler.Reconcile(xds_context.Context{}, &proxy)
		Expect(err).To(Not(Succeed()))
		Expect(err.Error()).To(ContainSubstring("not implemented"))

		snap, err := serverCtx.Cache().GetSnapshot(proxy.Id.String())
		Expect(err).To(Not(Succeed()))
		Expect(snap.Resources[envoy_types.Listener].Items).To(BeEmpty())
	})
})
