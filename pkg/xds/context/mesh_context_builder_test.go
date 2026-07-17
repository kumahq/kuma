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
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/destinationname"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
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

var _ = Describe("ServicesInformation", func() {
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

	It("resolves TLS readiness off MeshService status and treats external services as always ready", func() {
		// given a mesh with a PERMISSIVE mTLS backend
		meshBuilder := builders.Mesh().
			WithBuiltinMTLSBackend("ca-1").
			WithEnabledMTLSBackend("ca-1").
			WithPermissiveMTLSBackends()
		Expect(meshBuilder.Create(resourceStore)).To(Succeed())
		meshName := meshBuilder.Build().GetMeta().GetName()

		// and a MeshService-backed service whose status is TLS ready
		msBuilder := builders.MeshService().
			WithMesh(meshName).
			WithName("backend").
			WithDataplaneTagsSelectorKV(mesh_proto.ServiceTag, "backend").
			AddIntPort(80, 8080, core_meta.ProtocolHTTP).
			WithTLSStatus(meshservice_api.TLSReady)
		Expect(msBuilder.Create(resourceStore)).To(Succeed())
		ms := msBuilder.Build()

		// and a legacy external service, which is always considered TLS ready
		externalService := &core_mesh.ExternalServiceResource{
			Meta: &test_model.ResourceMeta{Mesh: meshName, Name: "external-svc"},
			Spec: &mesh_proto.ExternalService{
				Networking: &mesh_proto.ExternalService_Networking{
					Address: "httpbin.org:80",
				},
				Tags: map[string]string{mesh_proto.ServiceTag: "external-svc"},
			},
		}
		Expect(resourceStore.Create(context.Background(), externalService, store.CreateByKey("external-svc", meshName))).To(Succeed())

		// and a builtin and a delegated gateway dataplane, neither of which is a regular service
		builtinGatewayBuilder := builders.Dataplane().
			WithMesh(meshName).
			WithName("gateway-builtin-dp").
			WithAddress("127.0.0.1").
			WithBuiltInGateway("gateway-builtin")
		Expect(builtinGatewayBuilder.Create(resourceStore)).To(Succeed())

		delegatedGatewayBuilder := builders.Dataplane().
			WithMesh(meshName).
			WithName("gateway-delegated-dp").
			WithAddress("127.0.0.1").
			WithBuiltInGateway("gateway-delegated")
		delegatedGateway := delegatedGatewayBuilder.Build()
		delegatedGateway.Spec.Networking.Gateway.Type = mesh_proto.Dataplane_Networking_Gateway_DELEGATED
		Expect(resourceStore.Create(context.Background(), delegatedGateway, store.CreateByKey("gateway-delegated-dp", meshName))).To(Succeed())

		// when
		mc, err := meshContextBuilder.Build(context.Background(), meshName)
		Expect(err).ToNot(HaveOccurred())

		// then the MeshService-backed service is TLS ready
		msKey := destinationname.MustResolve(false, ms, ms.Spec.Ports[0])
		Expect(mc.ServicesInformation[msKey]).ToNot(BeNil())
		Expect(mc.ServicesInformation[msKey].TLSReadiness).To(BeTrue())

		// and the external service is unconditionally TLS ready
		Expect(mc.ServicesInformation["external-svc"]).ToNot(BeNil())
		Expect(mc.ServicesInformation["external-svc"].IsExternalService).To(BeTrue())
		Expect(mc.ServicesInformation["external-svc"].TLSReadiness).To(BeTrue())

		// and gateway dataplanes (builtin and delegated) never had ServiceInsight-backed
		// TLS readiness and still don't get a ServicesInformation entry of their own
		Expect(mc.ServicesInformation).ToNot(HaveKey("gateway-builtin"))
		Expect(mc.ServicesInformation).ToNot(HaveKey("gateway-delegated"))
	})

	It("does not mark services TLS ready when the mesh CA backend is not PERMISSIVE", func() {
		meshBuilder := builders.Mesh().
			WithBuiltinMTLSBackend("ca-1").
			WithEnabledMTLSBackend("ca-1")
		Expect(meshBuilder.Create(resourceStore)).To(Succeed())
		meshName := meshBuilder.Build().GetMeta().GetName()

		msBuilder := builders.MeshService().
			WithMesh(meshName).
			WithName("backend").
			WithDataplaneTagsSelectorKV(mesh_proto.ServiceTag, "backend").
			AddIntPort(80, 8080, core_meta.ProtocolHTTP).
			WithTLSStatus(meshservice_api.TLSReady)
		Expect(msBuilder.Create(resourceStore)).To(Succeed())
		ms := msBuilder.Build()

		mc, err := meshContextBuilder.Build(context.Background(), meshName)
		Expect(err).ToNot(HaveOccurred())

		msKey := destinationname.MustResolve(false, ms, ms.Spec.Ports[0])
		info, found := mc.ServicesInformation[msKey]
		if found {
			Expect(info.TLSReadiness).To(BeFalse())
		}
	})
})
