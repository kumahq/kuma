package inspect_test

import (
	"context"
	"encoding/json"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds/inspect"
	"github.com/kumahq/kuma/pkg/dns/vips"
	meshtimeout_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/server"
)

var _ = Describe("ProxyConfigInspector", func() {
	zone := "zone-1"
	mesh := "mesh-1"

	var resManager manager.ResourceManager
	var meshContextBuilder xds_context.MeshContextBuilder

	BeforeEach(func() {
		store := memory.NewStore()
		resManager = manager.NewResourceManager(store)

		meshContextBuilder = xds_context.NewMeshContextBuilder(
			resManager,
			server.MeshResourceTypes(),
			net.LookupIP,
			zone,
			vips.NewPersistence(resManager, config_manager.NewConfigManager(store), false),
			".mesh",
			80,
			xds_context.AnyToAnyReachableServicesGraphBuilder,
		)

		Expect(resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(mesh, model.NoMesh))).To(Succeed())
	})

	createDPP := func(name string) {
		dataplane := builders.Dataplane().
			WithAddress("192.168.0.1").
			WithInboundOfTags(mesh_proto.ServiceTag, "test", "region", "eu").
			Build()
		Expect(resManager.Create(context.Background(), dataplane, core_store.CreateByKey(name, mesh))).To(Succeed())
	}

	createMeshTimeout := func(name string, shadow bool) {
		mt := builders.MeshTimeout().
			WithMesh(mesh).WithName(name).
			WithTargetRef(builders.TargetRefMesh()).
			AddFrom(builders.TargetRefMesh(), meshtimeout_api.Conf{
				IdleTimeout: &kube_meta.Duration{Duration: 123 * time.Second},
			}).
			Build()

		labels := map[string]string{}
		if shadow {
			labels[mesh_proto.EffectLabel] = "shadow"
		}
		Expect(resManager.Create(context.Background(), mt, core_store.CreateByKey(name, mesh), core_store.CreateWithLabels(labels))).To(Succeed())
	}

	It("should return a ProxyConfig", func() {
		// given
		createDPP("test-dpp")
		createMeshTimeout("mt-1", false)

		meshContext, err := meshContextBuilder.Build(context.Background(), mesh)
		Expect(err).ToNot(HaveOccurred())
		inspector, err := inspect.NewProxyConfigInspector(meshContext, zone)
		Expect(err).ToNot(HaveOccurred())

		// when
		config, err := inspector.Get(context.Background(), "test-dpp", false)
		Expect(err).ToNot(HaveOccurred())

		// then
		bytes, _ := json.MarshalIndent(config, "", "  ")
		Expect(bytes).To(matchers.MatchGoldenJSON("testdata", "no-shadow-policies.get.json"))
	})

	It("should return a ProxyConfig with shadow", func() {
		// given
		createDPP("test-dpp")
		createMeshTimeout("mt-1", true)

		meshContext, err := meshContextBuilder.Build(context.Background(), mesh)
		Expect(err).ToNot(HaveOccurred())
		inspector, err := inspect.NewProxyConfigInspector(meshContext, zone)
		Expect(err).ToNot(HaveOccurred())

		// when
		config, err := inspector.Get(context.Background(), "test-dpp", false)
		Expect(err).ToNot(HaveOccurred())

		// then
		bytes, _ := json.MarshalIndent(config, "", "  ")
		Expect(bytes).To(matchers.MatchGoldenJSON("testdata", "with-shadow-policies.get.json"))

		// and when
		config, err = inspector.Get(context.Background(), "test-dpp", true)
		Expect(err).ToNot(HaveOccurred())

		// then
		bytes, _ = json.MarshalIndent(config, "", "  ")
		Expect(bytes).To(matchers.MatchGoldenJSON("testdata", "with-shadow-policies.get-shadow.json"))
	})

	It("should return a JSONPatch diff between proxy configs", func() {
		// given
		createDPP("test-dpp")
		createMeshTimeout("mt-1", true)

		meshContext, err := meshContextBuilder.Build(context.Background(), mesh)
		Expect(err).ToNot(HaveOccurred())
		inspector, err := inspect.NewProxyConfigInspector(meshContext, zone)
		Expect(err).ToNot(HaveOccurred())

		noShadow, err := inspector.Get(context.Background(), "test-dpp", false)
		Expect(err).ToNot(HaveOccurred())
		shadow, err := inspector.Get(context.Background(), "test-dpp", true)
		Expect(err).ToNot(HaveOccurred())

		// when
		diff, err := inspect.Diff(shadow, noShadow)
		Expect(err).ToNot(HaveOccurred())

		bytes, _ := json.MarshalIndent(diff, "", "  ")
		Expect(bytes).To(matchers.MatchGoldenJSON("testdata", "diff.json"))
	})
})
