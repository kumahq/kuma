package context_test

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v2/pkg/core/config/manager"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_manager "github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/dns/vips"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	test_store "github.com/kumahq/kuma/v2/pkg/test/store"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	xds_server "github.com/kumahq/kuma/v2/pkg/xds/server"
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
			vips.NewPersistence(core_manager.NewResourceManager(resourceStore), manager.NewConfigManager(resourceStore), false),
			"mesh",
			80,
			xds_context.AnyToAnyReachableServicesGraphBuilder,
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

	It("resolves a Dataplane address that is a DNS name", func() {
		// given a builder whose lookup resolves a hostname, as it would on AWS ECS
		// where the Dataplane is created before the container gets its IP
		builderWithDNS := xds_context.NewMeshContextBuilder(
			resourceStore,
			xds_server.MeshResourceTypes(),
			func(host string) ([]net.IP, error) {
				switch host {
				case "backend.dns.name":
					return []net.IP{net.ParseIP("192.168.0.10")}, nil
				case "backend.advertised.dns.name":
					return []net.IP{net.ParseIP("192.168.0.11")}, nil
				default:
					return []net.IP{net.ParseIP(host)}, nil
				}
			},
			"zone-1",
			vips.NewPersistence(core_manager.NewResourceManager(resourceStore), manager.NewConfigManager(resourceStore), false),
			"mesh",
			80,
			xds_context.AnyToAnyReachableServicesGraphBuilder,
			nil,
		)

		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Mesh
name: mesh-1
---
type: Dataplane
name: dp-1
mesh: mesh-1
networking:
  address: backend.dns.name
  advertisedAddress: backend.advertised.dns.name
  inbound:
    - port: 8080
      tags:
        kuma.io/service: backend
`)).To(Succeed())

		// when
		meshCtx, err := builderWithDNS.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())

		// then the dataplane stored in the mesh context is resolved
		dataplanes := meshCtx.Resources.Dataplanes().Items
		Expect(dataplanes).To(HaveLen(1))
		Expect(dataplanes[0].Spec.GetNetworking().GetAddress()).To(Equal("192.168.0.10"))
		Expect(dataplanes[0].Spec.GetNetworking().GetAdvertisedAddress()).To(Equal("192.168.0.11"))

		// and so is the endpoint Envoy EDS gets, since it only accepts IPs
		Expect(meshCtx.EndpointMap["backend"]).To(HaveLen(1))
		Expect(meshCtx.EndpointMap["backend"][0].Target).To(Equal("192.168.0.11"))
	})

	It("returns an error instead of panicking when listing resources fails", func() {
		// given a mesh exists but listing any other resource fails (e.g. DB connection lost).
		// MeshService is a destination type, whose fetched list used to be read before the
		// error check - that is where the nil-pointer panic happened
		Expect(samples.MeshDefaultBuilder().Create(resourceStore)).To(Succeed())
		builder := xds_context.NewMeshContextBuilder(
			&failingListManager{ReadOnlyResourceManager: resourceStore},
			[]core_model.ResourceType{meshservice_api.MeshServiceType},
			lookupIPFunc,
			"zone-1",
			vips.NewPersistence(core_manager.NewResourceManager(resourceStore), manager.NewConfigManager(resourceStore), false),
			"mesh",
			80,
			xds_context.AnyToAnyReachableServicesGraphBuilder,
			nil,
		)

		// when building the base mesh context
		_, err := builder.BuildBaseMeshContextIfChanged(context.Background(), "default", nil)

		// then the error is surfaced rather than causing a nil-pointer panic
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to build base mesh context"))
	})
})

// failingListManager delegates Get to a real manager but fails every List, simulating a store error.
type failingListManager struct {
	core_manager.ReadOnlyResourceManager
}

func (f *failingListManager) List(context.Context, core_model.ResourceList, ...store.ListOptionsFunc) error {
	return errors.New("store unavailable")
}
