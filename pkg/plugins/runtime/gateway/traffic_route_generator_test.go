package gateway_test

import (
	"fmt"
	"path"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/test/matchers"
	xds_server "github.com/kumahq/kuma/pkg/xds/server/v3"
)

var _ = Describe("Gateway Traffic Route", func() {
	var rt runtime.Runtime
	var addressCount int

	Do := func() (cache.Snapshot, error) {
		serverCtx := xds_server.NewXdsContext()
		reconciler := xds_server.DefaultReconciler(rt, serverCtx)

		// We expect there to be a Dataplane fixture named
		// "default" in the current mesh.
		ctx, proxy := MakeGeneratorContext(rt,
			core_model.ResourceKey{Mesh: "default", Name: "default"})

		Expect(proxy.Dataplane.Spec.IsBuiltinGateway()).To(BeTrue())

		if err := reconciler.Reconcile(*ctx, proxy); err != nil {
			return cache.Snapshot{}, err
		}

		return serverCtx.Cache().GetSnapshot(proxy.Id.String())
	}

	CreateServiceDataplanes := func(serviceName string, count int, tags ...string) {
		for i := 0; i < count; i++ {
			dp := mesh.NewDataplaneResource()
			dp.SetMeta(&rest.ResourceMeta{
				Type:             string(dp.Descriptor().Name),
				Mesh:             "default",
				Name:             fmt.Sprintf("%s-%d", serviceName, i),
				CreationTime:     core.Now(),
				ModificationTime: core.Now(),
			})
			dp.Spec.Networking = &mesh_proto.Dataplane_Networking{
				Address: fmt.Sprintf("192.168.1.%d", addressCount+1),
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					&mesh_proto.Dataplane_Networking_Inbound{
						Port:        20011,
						ServicePort: 10011,
						Tags: map[string]string{
							mesh_proto.ServiceTag:  serviceName,
							mesh_proto.ProtocolTag: "http",
						},
					}},
			}

			// Attach tag pairs. It's OK to panic if caller didn't pass even pairs.
			for i := 0; i < len(tags); i = i + 2 {
				dp.Spec.GetNetworking().GetInbound()[0].Tags[tags[i]] = tags[i+1]
			}

			Expect(StoreFixture(rt, dp)).To(Succeed())
			addressCount++
		}
	}

	BeforeEach(func() {
		var err error

		addressCount = 0

		rt, err = BuildRuntime()
		Expect(err).To(Succeed(), "build runtime instance")

		Expect(StoreNamedFixture(rt, "mesh-default.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "dataplane-default.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "gateway-default.yaml")).To(Succeed())
	})

	It("should generate from traffic routes", func() {
		// given
		CreateServiceDataplanes("echo-service", 5)

		Expect(StoreInlineFixture(rt, []byte(`
type: TrafficRoute
mesh: default
name: echo-service
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: '*'
conf:
  http:
  - match:
      path:
        prefix: /
    destination:
      kuma.io/service: echo-service
  destination:
    kuma.io/service: echo-service
`))).To(Succeed())

		// when
		snap, err := Do()
		Expect(err).To(Succeed())

		// then
		Expect(yaml.Marshal(MakeProtoSnapshot(snap))).
			To(matchers.MatchGoldenYAML(path.Join("testdata", "01-traffic-route.yaml")))
	})

	It("should generate resources from multiple traffic routes", func() {
		// given
		CreateServiceDataplanes("echo-service", 2)

		Expect(StoreInlineFixture(rt, []byte(`
type: TrafficRoute
mesh: default
name: edge-gateway-1
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: '*'
conf:
  http:
  - match:
      path:
        prefix: /
    destination:
      kuma.io/service: echo-service
  destination:
    kuma.io/service: echo-service
`))).To(Succeed())

		Expect(StoreInlineFixture(rt, []byte(`
type: TrafficRoute
mesh: default
name: edge-gateway-2
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: '*'
conf:
  http:
  - match:
      path:
        prefix: /
    destination:
      kuma.io/service: echo-service
  destination:
    kuma.io/service: echo-service
`))).To(Succeed())

		// when
		snap, err := Do()
		Expect(err).To(Succeed())

		// then
		Expect(yaml.Marshal(MakeProtoSnapshot(snap))).
			To(matchers.MatchGoldenYAML(path.Join("testdata", "02-traffic-route.yaml")))
	})

	// This test creates multiple versions of the same service and splits them by tags. We want to ensure
	// that the clusters are unique even though the splits happen across different resources.
	It("should split clusters across multiple traffic routes", func() {
		// given
		CreateServiceDataplanes(
			"echo-service-one", 1,
			mesh_proto.ServiceTag, "echo-service",
			"version", "1",
		)

		CreateServiceDataplanes(
			"echo-service-two", 1,
			mesh_proto.ServiceTag, "echo-service",
			"version", "2",
		)

		CreateServiceDataplanes(
			"rumored-echo-service-one", 1,
			mesh_proto.ServiceTag, "echo-service",
			"rumored-version", "1",
		)

		CreateServiceDataplanes(
			"rumored-echo-service-two", 1,
			mesh_proto.ServiceTag, "echo-service",
			"rumored-version", "2",
		)

		Expect(StoreInlineFixture(rt, []byte(`
type: TrafficRoute
mesh: default
name: edge-gateway-1
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: '*'
conf:
  http:
  - match:
      path:
        prefix: /api
    split:
    - weight: 1
      destination:
        kuma.io/service: echo-service
        version: "1"
    - weight: 1
      destination:
        kuma.io/service: echo-service
        version: "2"
  - match:
      path:
        prefix: /api/foo
    destination:
      kuma.io/service: echo-service
      version: "2"
  destination:
    kuma.io/service: echo-service
`))).To(Succeed())

		Expect(StoreInlineFixture(rt, []byte(`
type: TrafficRoute
mesh: default
name: edge-gateway-2
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: '*'
conf:
  http:
  - match:
      path:
        prefix: /phoney
    split:
    - weight: 4
      destination:
        kuma.io/service: echo-service
        rumored-version: "1"
    - weight: 5
      destination:
        kuma.io/service: echo-service
        rumored-version: "2"
  - match:
      path:
        prefix: /honey
    destination:
      kuma.io/service: echo-service
      rumored-version: "2"
  destination:
    kuma.io/service: echo-service
`))).To(Succeed())

		// when
		snap, err := Do()
		Expect(err).To(Succeed())

		resources := MakeProtoSnapshot(snap)

		// There are 4 unique service+tags combinations for destinations.
		Expect(len(resources.Clusters.Resources)).Should(BeNumerically("==", 4))

		// then
		Expect(yaml.Marshal(MakeProtoSnapshot(snap))).
			To(matchers.MatchGoldenYAML(path.Join("testdata", "03-traffic-route.yaml")))
	})

	It("should generate multiple virtual hosts", func() {
		// given
		CreateServiceDataplanes("echo-service", 2)

		// Update the default gateway to have multiple listeners.
		Expect(StoreInlineFixture(rt, []byte(`
type: Gateway
mesh: default
name: edge-gateway
sources:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    hostname: foo.example.com
    tags:
      name: foo.example.com
  - port: 8080
    protocol: HTTP
    hostname: bar.example.com
    tags:
      name: bar.example.com
`))).To(Succeed())

		Expect(StoreInlineFixture(rt, []byte(`
type: TrafficRoute
mesh: default
name: foo-routes
sources:
- match:
    kuma.io/service: gateway-default
    name: foo.example.com
destinations:
- match:
    kuma.io/service: '*'
conf:
  http:
  - match:
      path:
        prefix: /api/foo
    destination:
      kuma.io/service: echo-service
  destination:
    kuma.io/service: echo-service
`))).To(Succeed())

		Expect(StoreInlineFixture(rt, []byte(`
type: TrafficRoute
mesh: default
name: bar-routes
sources:
- match:
    kuma.io/service: gateway-default
    name: bar.example.com
destinations:
- match:
    kuma.io/service: '*'
conf:
  http:
  - match:
      path:
        prefix: /api/bar
    destination:
      kuma.io/service: echo-service
  destination:
    kuma.io/service: echo-service
`))).To(Succeed())

		Expect(StoreInlineFixture(rt, []byte(`
type: TrafficRoute
mesh: default
name: common-routes
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: '*'
conf:
  http:
  - match:
      path:
        prefix: /healthz
    destination:
      kuma.io/service: echo-service
  destination:
    kuma.io/service: echo-service
`))).To(Succeed())

		// when
		snap, err := Do()
		Expect(err).To(Succeed())

		// then
		Expect(yaml.Marshal(MakeProtoSnapshot(snap))).
			To(matchers.MatchGoldenYAML(path.Join("testdata", "04-traffic-route.yaml")))
	})
})
