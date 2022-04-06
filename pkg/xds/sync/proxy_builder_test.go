package sync_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns/vips"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/server"
	"github.com/kumahq/kuma/pkg/xds/sync"
)

type mockMetadataTracker struct{}

func (m mockMetadataTracker) Metadata(dpKey core_model.ResourceKey) *core_xds.DataplaneMetadata {
	return nil
}

func ToYAML(proxy *core_xds.Proxy) ([]byte, error) {
	out, err := yaml.Marshal(proxy)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func initializeStore(ctx context.Context, resourceManager core_manager.ResourceManager, fileWithResourcesName string) {
	resourcePath := filepath.Join(
		"testdata", "input", fileWithResourcesName,
	)

	resourceBytes, err := os.ReadFile(resourcePath)
	Expect(err).ToNot(HaveOccurred())

	rawResources := strings.Split(string(resourceBytes), "---")
	for _, rawResource := range rawResources {
		bytes := []byte(rawResource)

		res, err := rest_types.UnmarshallToCore(bytes)
		Expect(err).To(BeNil())

		switch res.Descriptor().Name {
		case core_mesh.ZoneEgressType:
			zoneEgress := res.(*core_mesh.ZoneEgressResource)
			Expect(resourceManager.Create(ctx, zoneEgress, store.CreateBy(core_model.MetaToResourceKey(zoneEgress.GetMeta())))).To(Succeed())
		case core_mesh.ZoneIngressType:
			zoneIngress := res.(*core_mesh.ZoneIngressResource)
			Expect(resourceManager.Create(ctx, zoneIngress, store.CreateBy(core_model.MetaToResourceKey(zoneIngress.GetMeta())))).To(Succeed())
		case core_mesh.DataplaneType:
			dataplane := res.(*core_mesh.DataplaneResource)
			Expect(resourceManager.Create(ctx, dataplane, store.CreateBy(core_model.MetaToResourceKey(dataplane.GetMeta())))).To(Succeed())
		case core_mesh.MeshType:
			meshResource := res.(*core_mesh.MeshResource)
			Expect(resourceManager.Create(ctx, meshResource, store.CreateBy(core_model.MetaToResourceKey(meshResource.GetMeta())))).To(Succeed())
		case core_mesh.ExternalServiceType:
			externalService := res.(*core_mesh.ExternalServiceResource)
			Expect(resourceManager.Create(ctx, externalService, store.CreateBy(core_model.MetaToResourceKey(externalService.GetMeta())))).To(Succeed())
		}
	}
}

var _ = Describe("Proxy Builder", func() {
	core.Now = func() time.Time {
		now, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
		return now
	}
	tracker := mockMetadataTracker{}
	localZone := "zone-1"

	ctx := context.Background()
	config := kuma_cp.DefaultConfig()
	config.Multizone.Zone.Name = localZone
	builder, err := test_runtime.BuilderFor(ctx, config)
	Expect(err).ToNot(HaveOccurred())
	rt, err := builder.Build()
	Expect(err).ToNot(HaveOccurred())

	meshCtxBuilder := xds_context.NewMeshContextBuilder(
		rt.ReadOnlyResourceManager(),
		server.MeshResourceTypes(server.HashMeshExcludedResources),
		rt.LookupIP(),
		rt.Config().Multizone.Zone.Name,
		vips.NewPersistence(rt.ReadOnlyResourceManager(), rt.ConfigManager()),
		rt.Config().DNSServer.Domain,
	)
	metrics, err := core_metrics.NewMetrics("cache")
	Expect(err).ToNot(HaveOccurred())
	meshCache, err := mesh.NewCache(rt.Config().Store.Cache.ExpirationTime, meshCtxBuilder, metrics)
	Expect(err).ToNot(HaveOccurred())
	initializeStore(ctx, rt.ResourceManager(), "default_resources.yaml")

	Describe("Build() zone egress", func() {
		egressProxyBuilder := sync.DefaultEgressProxyBuilder(
			ctx,
			rt,
			tracker,
			meshCache,
			envoy_common.APIV3,
		)

		It("should build proxy object for egress", func() {
			// given
			rk := core_model.ResourceKey{Name: "zone-egress-1"}

			// when
			proxy, err := egressProxyBuilder.Build(rk)
			Expect(err).ToNot(HaveOccurred())
			proxyYaml, err := ToYAML(proxy)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(proxyYaml).To(matchers.MatchGoldenYAML(filepath.Join("testdata", "01.zone-egress-proxy.golden.yaml")))
		})
	})

	Describe("Build() zone ingress", func() {
		ingressProxyBuilder := sync.DefaultIngressProxyBuilder(
			rt,
			tracker,
			envoy_common.APIV3,
			meshCache,
		)

		It("should build proxy object for ingress", func() {
			// given
			rk := core_model.ResourceKey{Name: "zone-ingress-zone-1"}

			// when
			proxy, err := ingressProxyBuilder.Build(rk)
			Expect(err).ToNot(HaveOccurred())
			proxyYaml, err := ToYAML(proxy)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(proxyYaml).To(matchers.MatchGoldenYAML(filepath.Join("testdata", "02.zone-ingress-proxy.golden.yaml")))
		})
	})
})
