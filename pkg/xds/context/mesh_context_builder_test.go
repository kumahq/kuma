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
	"github.com/kumahq/kuma/v2/pkg/core/dns/lookup"
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

func newMeshContextBuilder(resourceStore store.ResourceStore, lookupIPFunc lookup.LookupIPFunc) xds_context.MeshContextBuilder {
	return xds_context.NewMeshContextBuilder(
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
}

var _ = Describe("hash", func() {
	lookupIPFunc := func(s string) ([]net.IP, error) {
		return []net.IP{net.ParseIP(s)}, nil
	}
	var resourceStore store.ResourceStore
	var meshContextBuilder xds_context.MeshContextBuilder

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		meshContextBuilder = newMeshContextBuilder(resourceStore, lookupIPFunc)
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

// remoteMeshZoneAddressWithHostname is the public address of the zone proxy in zone
// "east", published as a load balancer hostname instead of an IP, which is what the
// ingress Service reconciler produces on EKS.
const remoteMeshZoneAddressWithHostname = `
type: MeshZoneAddress
name: zone-proxy-east
mesh: default
labels:
  kuma.io/zone: east
  kuma.io/origin: global
spec:
  address: lb.example.com
  port: 10001
`

// remoteZoneIngress is the legacy zone proxy of zone "east", used whenever that zone
// publishes no usable MeshZoneAddress.
const remoteZoneIngress = `
type: ZoneIngress
name: ingress-east
zone: east
networking:
  address: 192.168.0.1
  port: 10001
  advertisedAddress: 20.0.0.1
  advertisedPort: 10001
`

func endpointTargets(meshCtx *xds_context.MeshContext) []string {
	var targets []string
	for _, endpoints := range meshCtx.EndpointMap {
		for _, endpoint := range endpoints {
			targets = append(targets, endpoint.Target)
		}
	}
	return targets
}

var _ = Describe("MeshZoneAddress", func() {
	var resourceStore store.ResourceStore

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		Expect(samples.MeshMTLSBuilder().Create(resourceStore)).To(Succeed())
		Expect(samples.MeshServiceSyncedBackendBuilder().Create(resourceStore)).To(Succeed())
		Expect(test_store.LoadResources(context.Background(), resourceStore, remoteMeshZoneAddressWithHostname)).To(Succeed())
	})

	It("resolves the MeshZoneAddress hostname before it reaches the endpoint map", func() {
		// given a MeshZoneAddress pointing at a load balancer hostname, as reconciled
		// from a LoadBalancer Service on EKS, where the hostname takes precedence
		meshContextBuilder := newMeshContextBuilder(resourceStore, func(host string) ([]net.IP, error) {
			if ip := net.ParseIP(host); ip != nil {
				return []net.IP{ip}, nil
			}
			return []net.IP{net.ParseIP("10.0.0.1")}, nil
		})

		// when
		meshCtx, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", nil)
		Expect(err).ToNot(HaveOccurred())

		// then the endpoint target is the resolved IP, not the hostname Envoy can't dial
		Expect(endpointTargets(meshCtx)).To(ConsistOf("10.0.0.1"))
		// and the zone stays classified as served by a mesh-scoped zone proxy, so consumers use the KRI SNI
		Expect(meshCtx.ZonesWithMeshScopedProxy).To(Equal(map[string]bool{"east": true}))
	})

	It("falls back to the ZoneIngress when the MeshZoneAddress hostname doesn't resolve", func() {
		// given the same zone also runs a legacy ZoneIngress, and the load balancer
		// hostname has no DNS record yet
		Expect(test_store.LoadResources(context.Background(), resourceStore, remoteZoneIngress)).To(Succeed())
		meshContextBuilder := newMeshContextBuilder(resourceStore, func(host string) ([]net.IP, error) {
			if ip := net.ParseIP(host); ip != nil {
				return []net.IP{ip}, nil
			}
			return nil, errors.New("no such host")
		})

		// when
		meshCtx, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", nil)
		Expect(err).ToNot(HaveOccurred())

		// then the unresolvable MeshZoneAddress is dropped and the zone keeps its ZoneIngress endpoint
		Expect(endpointTargets(meshCtx)).To(ConsistOf("20.0.0.1"))
		// and the zone leaves ZonesWithMeshScopedProxy together with its endpoint, so consumers
		// use the hash-based SNI that the legacy ZoneIngress they now dial actually serves
		Expect(meshCtx.ZonesWithMeshScopedProxy).To(BeEmpty())
	})
})

var _ = Describe("zone egress endpoints", func() {
	lookupIPFunc := func(s string) ([]net.IP, error) {
		return []net.IP{net.ParseIP(s)}, nil
	}

	It("skips zone egresses whose pod is terminating or not ready", func() {
		// given
		resourceStore := memory.NewStore()
		meshContextBuilder := newMeshContextBuilder(resourceStore, lookupIPFunc)
		Expect(test_store.LoadResources(context.Background(), resourceStore, `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
    - name: ca-1
      type: builtin
routing:
  zoneEgress: true
---
type: ExternalService
name: httpbin
mesh: default
networking:
  address: httpbin.org:80
tags:
  kuma.io/service: httpbin
---
type: ZoneEgress
name: egress-ready
labels:
  kuma.io/proxy-ready: "true"
networking:
  address: 192.168.0.1
  port: 10002
---
type: ZoneEgress
name: egress-terminating
labels:
  kuma.io/proxy-ready: "false"
networking:
  address: 192.168.0.2
  port: 10002
---
type: ZoneEgress
name: egress-universal
networking:
  address: 192.168.0.3
  port: 10002
`)).To(Succeed())

		// when
		meshContext, err := meshContextBuilder.BuildIfChanged(context.Background(), "default", nil)

		// then the terminating instance is gone, while the one with no label at all - a Universal egress or one
		// written by an older CP - is kept
		Expect(err).ToNot(HaveOccurred())
		var targets []string
		for _, endpoint := range meshContext.EndpointMap["httpbin"] {
			targets = append(targets, endpoint.Target)
		}
		Expect(targets).To(ConsistOf("192.168.0.1", "192.168.0.3"))

		// and the terminating instance is still in the mesh context, so that it keeps being served its own
		// configuration until it exits
		var names []string
		for _, ze := range meshContext.Resources.ZoneEgresses().Items {
			names = append(names, ze.GetMeta().GetName())
		}
		Expect(names).To(ConsistOf("egress-ready", "egress-terminating", "egress-universal"))

		// and it resolves by name, which is how the egress proxy builder finds it
		aggregated, err := xds_context.AggregateMeshContexts(
			context.Background(),
			core_manager.NewResourceManager(resourceStore),
			meshContextBuilder.Build,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(aggregated.ZoneEgressByName).To(HaveKey("egress-terminating"))
	})
})

// failingListManager delegates Get to a real manager but fails every List, simulating a store error.
type failingListManager struct {
	core_manager.ReadOnlyResourceManager
}

func (f *failingListManager) List(context.Context, core_model.ResourceList, ...store.ListOptionsFunc) error {
	return errors.New("store unavailable")
}
