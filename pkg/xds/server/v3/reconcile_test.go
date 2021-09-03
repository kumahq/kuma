package v3

import (
	"fmt"
	"sync/atomic"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	xds_model "github.com/kumahq/kuma/pkg/core/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ = Describe("Reconcile", func() {
	Describe("reconciler", func() {

		var backupNewUUID func() string

		BeforeEach(func() {
			backupNewUUID = core.NewUUID
		})
		AfterEach(func() {
			core.NewUUID = backupNewUUID
		})

		var serial uint64

		BeforeEach(func() {
			core.NewUUID = func() string {
				uuid := atomic.AddUint64(&serial, 1)
				return fmt.Sprintf("v%d", uuid)
			}
		})

		var xdsContext XdsContext

		BeforeEach(func() {
			xdsContext = NewXdsContext()
		})

		snapshot := envoy_cache.Snapshot{
			Resources: [envoy_types.UnknownType]envoy_cache.Resources{
				envoy_types.Listener: {
					Items: map[string]envoy_types.ResourceWithTtl{
						"listener": {
							Resource: &envoy_listener.Listener{},
						},
					},
				},
				envoy_types.Route: {
					Items: map[string]envoy_types.ResourceWithTtl{
						"route": {
							Resource: &envoy_route.RouteConfiguration{},
						},
					},
				},
				envoy_types.Cluster: {
					Items: map[string]envoy_types.ResourceWithTtl{
						"cluster": {
							Resource: &envoy_cluster.Cluster{},
						},
					},
				},
				envoy_types.Endpoint: {
					Items: map[string]envoy_types.ResourceWithTtl{
						"endpoint": {
							Resource: &envoy_endpoint.ClusterLoadAssignment{},
						},
					},
				},
				envoy_types.Secret: {
					Items: map[string]envoy_types.ResourceWithTtl{
						"secret": {
							Resource: &envoy_auth.Secret{},
						},
					},
				},
			},
		}

		It("should generate a Snapshot per Envoy Node", func() {
			// given
			snapshots := make(chan envoy_cache.Snapshot, 3)
			snapshots <- snapshot               // initial Dataplane configuration
			snapshots <- snapshot               // same Dataplane configuration
			snapshots <- envoy_cache.Snapshot{} // new Dataplane configuration

			// setup
			r := &reconciler{
				snapshotGeneratorFunc(func(ctx xds_context.Context, proxy *xds_model.Proxy) (envoy_cache.Snapshot, error) {
					return <-snapshots, nil
				}),
				&simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
			}

			// given
			dataplane := &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Mesh:    "demo",
					Name:    "example",
					Version: "abcdefg",
				},
				Spec: &mesh_proto.Dataplane{},
			}

			By("simulating discovery event")
			// when
			proxy := &xds_model.Proxy{
				Id:        *xds_model.BuildProxyId("demo", "example"),
				Dataplane: dataplane,
			}
			err := r.Reconcile(xds_context.Context{}, proxy)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("verifying that snapshot versions were auto-generated")
			// when
			snapshot, err := xdsContext.Cache().GetSnapshot("demo.example")
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshot).ToNot(BeZero())
			// and
			Expect(snapshot.Resources[envoy_types.Listener].Version).To(Equal("v1"))
			Expect(snapshot.Resources[envoy_types.Route].Version).To(Equal("v2"))
			Expect(snapshot.Resources[envoy_types.Cluster].Version).To(Equal("v3"))
			Expect(snapshot.Resources[envoy_types.Endpoint].Version).To(Equal("v4"))
			Expect(snapshot.Resources[envoy_types.Secret].Version).To(Equal("v5"))

			By("simulating discovery event (Dataplane watchdog triggers refresh)")
			// when
			err = r.Reconcile(xds_context.Context{}, proxy)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("verifying that snapshot versions remain the same")
			// when
			snapshot, err = xdsContext.Cache().GetSnapshot("demo.example")
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshot).ToNot(BeZero())
			// and
			Expect(snapshot.Resources[envoy_types.Listener].Version).To(Equal("v1"))
			Expect(snapshot.Resources[envoy_types.Route].Version).To(Equal("v2"))
			Expect(snapshot.Resources[envoy_types.Cluster].Version).To(Equal("v3"))
			Expect(snapshot.Resources[envoy_types.Endpoint].Version).To(Equal("v4"))
			Expect(snapshot.Resources[envoy_types.Secret].Version).To(Equal("v5"))

			By("simulating discovery event (Dataplane gets changed)")
			// when
			err = r.Reconcile(xds_context.Context{}, proxy)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("verifying that snapshot versions are new")
			// when
			snapshot, err = xdsContext.Cache().GetSnapshot("demo.example")
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshot).ToNot(BeZero())
			// and
			Expect(snapshot.Resources[envoy_types.Listener].Version).To(Equal("v6"))
			Expect(snapshot.Resources[envoy_types.Route].Version).To(Equal("v7"))
			Expect(snapshot.Resources[envoy_types.Cluster].Version).To(Equal("v8"))
			Expect(snapshot.Resources[envoy_types.Endpoint].Version).To(Equal("v9"))
			Expect(snapshot.Resources[envoy_types.Secret].Version).To(Equal("v10"))

			By("simulating clear")
			// when
			err = r.Clear(&proxy.Id)
			Expect(err).ToNot(HaveOccurred())
			snapshot, err = xdsContext.Cache().GetSnapshot("demo.example")

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no snapshot found"))

			Expect(snapshot.Resources[envoy_types.Listener].Version).To(Equal(""))
			Expect(snapshot.Resources[envoy_types.Route].Version).To(Equal(""))
			Expect(snapshot.Resources[envoy_types.Cluster].Version).To(Equal(""))
			Expect(snapshot.Resources[envoy_types.Endpoint].Version).To(Equal(""))
			Expect(snapshot.Resources[envoy_types.Secret].Version).To(Equal(""))
		})
	})
})

type snapshotGeneratorFunc func(ctx xds_context.Context, proxy *xds_model.Proxy) (envoy_cache.Snapshot, error)

func (f snapshotGeneratorFunc) GenerateSnapshot(ctx xds_context.Context, proxy *xds_model.Proxy) (envoy_cache.Snapshot, error) {
	return f(ctx, proxy)
}
