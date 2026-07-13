package context_test

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/config/manager"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/dns/vips"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test"
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
			vips.NewPersistence(core_manager.NewResourceManager(resourceStore), manager.NewConfigManager(resourceStore), false),
			"mesh",
			80,
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

	It("should not recompute the mesh context when a Dataplane write only bumps resourceVersion", func() {
		// given a mesh with a single Dataplane
		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Mesh
name: mesh-1
---
type: Dataplane
name: dp-1
mesh: mesh-1
networking:
  address: 127.0.0.1
  inbound:
    - port: 8080
      tags:
        kuma.io/service: backend
`)).To(Succeed())

		before, err := meshContextBuilder.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())

		// when the Dataplane is written again with an identical spec (only resourceVersion changes)
		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Dataplane
name: dp-1
mesh: mesh-1
networking:
  address: 127.0.0.1
  inbound:
    - port: 8080
      tags:
        kuma.io/service: backend
`)).To(Succeed())

		after, err := meshContextBuilder.BuildIfChanged(context.Background(), "mesh-1", before)
		Expect(err).ToNot(HaveOccurred())

		// then the cached context is reused and the mesh hash is unchanged
		Expect(after).To(BeIdenticalTo(before), "resourceVersion-only write should not trigger mesh-wide xDS recomputation")
		Expect(after.Hash).To(Equal(before.Hash))
	})

	It("should not recompute the mesh context when a MeshService write only bumps DataplaneProxies stats", func() {
		// given a mesh with a single MeshService
		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Mesh
name: mesh-1
---
type: MeshService
name: redis
mesh: mesh-1
spec:
  selector:
    dataplaneTags:
      app: redis
  ports:
  - port: 6739
    appProtocol: tcp
status:
  vips:
  - ip: 10.0.1.1
`)).To(Succeed())

		before, err := meshContextBuilder.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())

		// when the MeshService is written again with the same spec/status but updated proxy stats
		meshService := meshservice_api.NewMeshServiceResource()
		Expect(resourceStore.Get(context.Background(), meshService, store.GetByKey("redis", "mesh-1"))).To(Succeed())
		meshService.Status.DataplaneProxies = meshservice_api.DataplaneProxies{
			Connected: 3,
			Healthy:   2,
			Total:     3,
		}
		Expect(resourceStore.Update(context.Background(), meshService, store.UpdateWithLabels(meshService.GetMeta().GetLabels()))).To(Succeed())

		after, err := meshContextBuilder.BuildIfChanged(context.Background(), "mesh-1", before)
		Expect(err).ToNot(HaveOccurred())

		// then the cached context is reused and the mesh hash is unchanged
		Expect(after).To(BeIdenticalTo(before), "DataplaneProxies-only write should not trigger mesh-wide xDS recomputation")
		Expect(after.Hash).To(Equal(before.Hash))
	})

	It("keeps PolicyMatchingHash stable across resourceVersion-only Dataplane writes", func() {
		builderWithPolicyMatchingHash := xds_context.NewMeshContextBuilder(
			resourceStore,
			xds_server.MeshResourceTypes(),
			lookupIPFunc,
			"zone-1",
			vips.NewPersistence(core_manager.NewResourceManager(resourceStore), manager.NewConfigManager(resourceStore), false),
			"mesh",
			80,
			nil,
			xds_context.WithPolicyMatchingHash(),
		)

		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Mesh
name: mesh-1
---
type: Dataplane
name: dp-1
mesh: mesh-1
networking:
  address: 127.0.0.1
  inbound:
    - port: 8080
      tags:
        kuma.io/service: backend
`)).To(Succeed())

		before, err := builderWithPolicyMatchingHash.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())

		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Dataplane
name: dp-1
mesh: mesh-1
networking:
  address: 127.0.0.1
  inbound:
    - port: 8080
      tags:
        kuma.io/service: backend
`)).To(Succeed())

		after, err := builderWithPolicyMatchingHash.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(after.PolicyMatchingHash).To(Equal(before.PolicyMatchingHash))
	})

	It("keeps PolicyMatchingHash stable across MeshService DataplaneProxies updates", func() {
		builderWithPolicyMatchingHash := xds_context.NewMeshContextBuilder(
			resourceStore,
			xds_server.MeshResourceTypes(),
			lookupIPFunc,
			"zone-1",
			vips.NewPersistence(core_manager.NewResourceManager(resourceStore), manager.NewConfigManager(resourceStore), false),
			"mesh",
			80,
			nil,
			xds_context.WithPolicyMatchingHash(),
		)

		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Mesh
name: mesh-1
---
type: MeshService
name: redis
mesh: mesh-1
spec:
  selector:
    dataplaneTags:
      app: redis
  ports:
  - port: 6739
    appProtocol: tcp
status:
  vips:
  - ip: 10.0.1.1
`)).To(Succeed())

		before, err := builderWithPolicyMatchingHash.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())

		meshService := meshservice_api.NewMeshServiceResource()
		Expect(resourceStore.Get(context.Background(), meshService, store.GetByKey("redis", "mesh-1"))).To(Succeed())
		meshService.Status.DataplaneProxies = meshservice_api.DataplaneProxies{
			Connected: 3,
			Healthy:   2,
			Total:     3,
		}
		Expect(resourceStore.Update(context.Background(), meshService, store.UpdateWithLabels(meshService.GetMeta().GetLabels()))).To(Succeed())

		after, err := builderWithPolicyMatchingHash.BuildIfChanged(context.Background(), "mesh-1", nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(after.PolicyMatchingHash).To(Equal(before.PolicyMatchingHash))
	})
})
