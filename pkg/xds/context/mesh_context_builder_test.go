package context_test

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	hostnamegenerator_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	system_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_store "github.com/kumahq/kuma/v3/pkg/test/store"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	xds_server "github.com/kumahq/kuma/v3/pkg/xds/server"
)

var _ = Describe("hash", func() {
	lookupIPFunc := func(s string) ([]net.IP, error) {
		return []net.IP{net.ParseIP(s)}, nil
	}
	var resourceStore store.ResourceStore
	var meshContextBuilder xds_context.MeshContextBuilder

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		meshContextBuilder = xds_context.NewMeshContextBuilder(
			resourceStore,
			xds_server.MeshResourceTypes(),
			lookupIPFunc,
			"zone-1",
			nil,
		)
	})

	_ = DescribeTable("with BaseMeshContext", func(inputFile string) {
		// Takes input.yaml compute the hash and then apply a set of changes, then check whether or not it stays the same
		// Given

		inputs, err := os.ReadFile(inputFile)
		Expect(err).NotTo(HaveOccurred())
		parts := strings.SplitN(string(inputs), "\n", 2)
		Expect(parts[0]).To(HavePrefix("#"), "the first line of the input is not a comment with the url path")
		actions := strings.Split(strings.Trim(parts[0], "# "), " ")
		Expect(actions).To(HaveLen(2), "the first line of the input should be: # <mesh> <bool to indicate if there a change or not>")
		Expect(test_store.LoadResources(context.Background(), resourceStore, string(inputs))).To(Succeed())

		meshName := strings.TrimSpace(actions[0])
		shouldChange, err := strconv.ParseBool(actions[1])
		Expect(err).ToNot(HaveOccurred())
		beforeContext, err := meshContextBuilder.BuildBaseMeshContextIfChanged(context.Background(), meshName, nil)
		Expect(err).ToNot(HaveOccurred())

		// When
		Expect(test_store.LoadResourcesFromFile(context.Background(), resourceStore, strings.Replace(inputFile, "input", "change", 1))).To(Succeed())

		// Then
		afterContext, err := meshContextBuilder.BuildBaseMeshContextIfChanged(context.Background(), meshName, beforeContext)
		Expect(err).ToNot(HaveOccurred())
		if shouldChange {
			Expect(afterContext.Hash()).ToNot(Equal(beforeContext.Hash()), "context didn't change when it should have")
			Expect(afterContext).ToNot(Equal(beforeContext))
		} else {
			Expect(afterContext.Hash()).To(Equal(beforeContext.Hash()), "context changed when it shouldn't have")
			Expect(afterContext).To(Equal(beforeContext), "context should be the exact same object")
		}
	}, test.EntriesForFolder("basemeshcontext_hash"))

	_ = DescribeTable("with GlobalContext", func(inputFile string) {
		// Takes input.yaml compute the hash and then apply a set of changes, then check whether or not it stays the same
		// Given

		inputs, err := os.ReadFile(inputFile)
		Expect(err).NotTo(HaveOccurred())
		parts := strings.SplitN(string(inputs), "\n", 2)
		Expect(parts[0]).To(HavePrefix("#"), "the first line of the input is not a comment with the url path")
		actions := strings.Split(strings.Trim(parts[0], "# "), " ")
		Expect(actions).To(HaveLen(1), "the first line of the input should be: # <bool to indicate if there a change or not>")
		Expect(test_store.LoadResources(context.Background(), resourceStore, string(inputs))).To(Succeed())

		shouldChange, err := strconv.ParseBool(actions[0])
		Expect(err).ToNot(HaveOccurred())

		beforeContext, err := meshContextBuilder.BuildGlobalContextIfChanged(context.Background(), nil)
		Expect(err).ToNot(HaveOccurred())

		// When
		Expect(test_store.LoadResourcesFromFile(context.Background(), resourceStore, strings.Replace(inputFile, "input", "change", 1))).To(Succeed())

		// Then
		afterContext, err := meshContextBuilder.BuildGlobalContextIfChanged(context.Background(), beforeContext)
		Expect(err).ToNot(HaveOccurred())
		if shouldChange {
			Expect(afterContext.Hash()).ToNot(Equal(beforeContext.Hash()), "context didn't change when it should have")
			Expect(afterContext).ToNot(Equal(beforeContext))
		} else {
			Expect(afterContext.Hash()).To(Equal(beforeContext.Hash()), "context changed when it shouldn't have")
			Expect(afterContext).To(Equal(beforeContext), "context should be the exact same object")
		}
	}, test.EntriesForFolder("globalcontext_hash"))

	_ = DescribeTable("with MeshContext", func(inputFile string) {
		// Takes input.yaml compute the hash and then apply a set of changes, then check whether or not it stays the same
		// Given

		inputs, err := os.ReadFile(inputFile)
		Expect(err).NotTo(HaveOccurred())
		parts := strings.SplitN(string(inputs), "\n", 2)
		Expect(parts[0]).To(HavePrefix("#"), "the first line of the input is not a comment with the url path")
		actions := strings.Split(strings.Trim(parts[0], "# "), " ")
		Expect(actions).To(HaveLen(2), "the first line of the input should be: # <mesh> <bool to indicate if there a change or not>")
		Expect(test_store.LoadResources(context.Background(), resourceStore, string(inputs))).To(Succeed())

		meshName := strings.TrimSpace(actions[0])
		shouldChange, err := strconv.ParseBool(actions[1])
		Expect(err).ToNot(HaveOccurred())
		beforeContext, err := meshContextBuilder.BuildIfChanged(context.Background(), meshName, nil)
		Expect(err).ToNot(HaveOccurred())

		// When
		Expect(test_store.LoadResourcesFromFile(context.Background(), resourceStore, strings.Replace(inputFile, "input", "change", 1))).To(Succeed())

		// Then
		afterContext, err := meshContextBuilder.BuildIfChanged(context.Background(), meshName, beforeContext)
		Expect(err).ToNot(HaveOccurred())
		if shouldChange {
			Expect(afterContext.Hash).ToNot(Equal(beforeContext.Hash), "context didn't change when it should have")
			Expect(afterContext).ToNot(Equal(beforeContext))
		} else {
			Expect(afterContext.Hash).To(Equal(beforeContext.Hash), "context changed when it shouldn't have")
			Expect(afterContext).To(Equal(beforeContext), "context should be the exact same object")
		}
	}, test.EntriesForFolder("meshcontext_hash"))
})

var _ = Describe("legacy VIP compatibility", func() {
	It("should ignore persisted legacy service VIPs when mesh services are exclusive", func() {
		ctx := context.Background()
		resourceStore := memory.NewStore()
		meshContextBuilder := xds_context.NewMeshContextBuilder(
			resourceStore,
			xds_server.MeshResourceTypes(),
			func(s string) ([]net.IP, error) {
				return []net.IP{net.ParseIP(s)}, nil
			},
			"zone-1",
			nil,
		)

		Expect(builders.Mesh().
			WithName("default").
			WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive).
			Create(resourceStore),
		).To(Succeed())

		meshService := builders.MeshService().
			WithMesh("default").
			WithName("local-test-server").
			WithNamespace("mesh-service-reachable-backends").
			WithLabels(map[string]string{
				mesh_proto.EnvTag:           mesh_proto.KubernetesEnvironment,
				mesh_proto.KubeNamespaceTag: "mesh-service-reachable-backends",
			}).
			AddIntPort(80, 80, "").
			WithKumaVIP("241.0.0.10").
			Build()
		meshService.Status.Addresses = []hostnamegenerator_api.Address{{
			Hostname: "local-test-server.mesh-service-reachable-backends.svc.kuma-1.mesh.local",
		}}
		Expect(resourceStore.Create(
			ctx,
			meshService,
			store.CreateByKey(meshService.GetMeta().GetName(), meshService.GetMeta().GetMesh()),
			store.CreateWithLabels(meshService.GetMeta().GetLabels()),
		)).To(Succeed())

		config := system_api.NewConfigResource()
		config.Spec = &system_proto.Config{
			Config: `{"0:local-test-server_mesh-service-reachable-backends_svc_80":{"address":"240.0.0.10","outbounds":[{"TagSet":{"kuma.io/service":"local-test-server_mesh-service-reachable-backends_svc_80"}}]}}`,
		}
		Expect(resourceStore.Create(
			ctx,
			config,
			store.CreateByKey("kuma-default-dns-vips", core_model.NoMesh),
		)).To(Succeed())

		meshCtx, err := meshContextBuilder.Build(ctx, "default")
		Expect(err).ToNot(HaveOccurred())

		Expect(meshCtx.VIPDomains).To(ContainElement(HaveField("Address", "241.0.0.10")))
		for _, vipDomain := range meshCtx.VIPDomains {
			Expect(vipDomain.Address).ToNot(Equal("240.0.0.10"))
		}
		for _, outbound := range meshCtx.VIPOutbounds {
			Expect(outbound.GetAddress()).ToNot(Equal("240.0.0.10"))
		}
	})
})
