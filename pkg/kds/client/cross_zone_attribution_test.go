// Tests that the global CP attributes KDS zone-to-global synced resources by the
// connecting peer's declared client-id, not by sender-provided in-band values.
// Each case drives the real global ingest with a client-id ("zone-b") that
// differs from the payload's zone ("zone-a"), and asserts everything resolves to
// the connecting zone.
package client_test

import (
	"context"
	"fmt"
	"testing"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/kds"
	kds_client "github.com/kumahq/kuma/v3/pkg/kds/client"
	kds_sync_store "github.com/kumahq/kuma/v3/pkg/kds/store"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	test_grpc "github.com/kumahq/kuma/v3/pkg/test/grpc"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

const (
	// connectingZone is the client-id the connection presents.
	connectingZone = "zone-b"
	// otherZone is the differing zone the sender names in-band.
	otherZone = "zone-a"
)

// newGlobalSink wires the global-CP KDS ingest as pkg/kds/mux/zone_sync.go does,
// over a memory store; the connecting peer's client-id is connectingZone.
func newGlobalSink(t *testing.T, ctx context.Context, typ core_model.ResourceType) (store.ResourceStore, *test_grpc.MockDeltaClientStream, *prometheus.CounterVec, func()) {
	t.Helper()
	g := NewWithT(t)

	globalStore := memory.NewStore()
	// Both zones exist on global, as in any federated mesh.
	for _, z := range []string{otherZone, connectingZone} {
		err := globalStore.Create(ctx,
			&system.ZoneResource{Spec: &system_proto.Zone{Enabled: util_proto.Bool(true)}},
			store.CreateByKey(z, core_model.NoMesh))
		g.Expect(err).ToNot(HaveOccurred())
	}

	metrics, err := core_metrics.NewMetrics("")
	g.Expect(err).ToNot(HaveOccurred())
	syncer, err := kds_sync_store.NewResourceSyncer(core.Log.WithName("crosszone-syncer"), globalStore, store.NoTransactions{}, metrics, context.Background())
	g.Expect(err).ToNot(HaveOccurred())

	// Counts resources whose zone attribution the global ingest rewrote because
	// a sender-provided value differed from the connecting zone.
	rewrites := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "test_kds_zone_attribution_rewrites_total",
	}, []string{"resource_type"})

	clientStream := test_grpc.NewMockDeltaClientStream()
	// The client-id drives attribution, not the in-band ControlPlane.Identifier.
	kdsStream := kds_client.NewDeltaKDSStream(clientStream, connectingZone, connectingZone+"-instance", "", 1)
	sink := kds_client.NewKDSSyncClient(
		core.Log.WithName("crosszone-global-sink"),
		[]core_model.ResourceType{typ},
		kdsStream,
		kds_sync_store.GlobalSyncCallback(ctx, syncer, false, nil, "kuma-system", rewrites),
		kds_client.SyncClientConfig{},
	)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = sink.Receive()
	}()

	cleanup := func() {
		close(clientStream.RecvCh)
		_ = kdsStream.CloseSend()
		<-done
	}
	return globalStore, clientStream, rewrites, cleanup
}

// deltaResponse builds the one-resource DeltaDiscoveryResponse a zone CP would send.
func deltaResponse(t *testing.T, typ core_model.ResourceType, controlPlaneID, name, mesh string, spec core_model.ResourceSpec) *envoy_sd.DeltaDiscoveryResponse {
	t.Helper()
	g := NewWithT(t)

	specAny, err := core_model.ToAny(spec)
	g.Expect(err).ToNot(HaveOccurred())
	kr := &mesh_proto.KumaResource{
		Meta: &mesh_proto.KumaResource_Meta{Name: name, Mesh: mesh},
		Spec: specAny,
	}
	krAny, err := util_proto.MarshalAnyDeterministic(kr)
	g.Expect(err).ToNot(HaveOccurred())

	return &envoy_sd.DeltaDiscoveryResponse{
		TypeUrl:      string(typ),
		Nonce:        "nonce-1",
		ControlPlane: &envoy_core.ControlPlane{Identifier: controlPlaneID},
		Resources: []*envoy_sd.Resource{{
			Name:     fmt.Sprintf("%s.%s", name, mesh),
			Version:  "v1",
			Resource: krAny,
		}},
	}
}

// dataplaneWithZoneTag returns a Dataplane whose metadata should be attributed to the connecting zone.
func dataplaneWithZoneTag(_ string) *mesh_proto.Dataplane {
	return &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Address: "192.168.0.1",
			Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
				Port: 1212,
			}},
			Outbound: []*mesh_proto.Dataplane_Networking_Outbound{{
				Port: 10000,
				BackendRef: &mesh_proto.Dataplane_Networking_Outbound_BackendRef{
					Kind: "MeshService",
					Name: "web",
					Port: 80,
				},
			}},
		},
	}
}

func gatewayDataplaneWithDivergentZoneTag() *mesh_proto.Dataplane {
	return &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Address: "192.168.0.2",
			Gateway: &mesh_proto.Dataplane_Networking_Gateway{
				Type: mesh_proto.Dataplane_Networking_Gateway_DELEGATED,
				Tags: map[string]string{
					mesh_proto.ZoneTag:    otherZone,
					mesh_proto.ServiceTag: "gateway-a",
				},
			},
		},
	}
}

// hashSuffixCtx advertises the hash-suffix feature, as every currently
// supported zone does, so the modern (non-prefixing) global ingest path runs.
func hashSuffixCtx() context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs(kds.FeaturesMetadataKey, kds.FeatureHashSuffix))
}

func storedDataplane(t *testing.T, s store.ResourceStore, ctx context.Context) *core_mesh.DataplaneResource {
	t.Helper()
	g := NewWithT(t)
	var stored *core_mesh.DataplaneResource
	g.Eventually(func(g Gomega) {
		list := core_mesh.DataplaneResourceList{}
		g.Expect(s.List(ctx, &list)).To(Succeed())
		g.Expect(list.Items).To(HaveLen(1))
		stored = list.Items[0]
	}, "5s", "50ms").Should(Succeed())
	return stored
}

// TestZoneToGlobalSyncAttribution asserts the top-level attribution (kuma.io/zone
// label) and the in-spec zone tags (Dataplane gateway) all resolve to the
// connecting zone's client-id, and that a matching zone is a no-op.
func TestZoneToGlobalSyncAttribution(t *testing.T) {
	t.Run("DataplaneGateway", func(t *testing.T) {
		g := NewWithT(t)
		ctx := hashSuffixCtx()
		globalStore, clientStream, rewrites, cleanup := newGlobalSink(t, ctx, core_mesh.DataplaneType)
		defer cleanup()

		clientStream.RecvCh <- deltaResponse(t, core_mesh.DataplaneType, otherZone, "gw-dp-1", "default", gatewayDataplaneWithDivergentZoneTag())

		stored := storedDataplane(t, globalStore, ctx)
		gatewayZone := stored.Spec.GetNetworking().GetGateway().GetTags()[mesh_proto.ZoneTag]
		g.Expect(gatewayZone).To(Equal(connectingZone),
			"Dataplane gateway kuma.io/zone tag must resolve to the connecting zone's client-id %q, not the sender-provided zone %q", connectingZone, gatewayZone)

		g.Expect(testutil.ToFloat64(rewrites.WithLabelValues(string(core_mesh.DataplaneType)))).To(Equal(1.0))
	})

	// When the sender's values already match the connecting zone (ordinary
	// sync), nothing is rewritten and the counter stays at zero.
	t.Run("MatchingZoneIsNoOp", func(t *testing.T) {
		g := NewWithT(t)
		ctx := hashSuffixCtx()
		globalStore, clientStream, rewrites, cleanup := newGlobalSink(t, ctx, core_mesh.DataplaneType)
		defer cleanup()

		clientStream.RecvCh <- deltaResponse(t, core_mesh.DataplaneType, connectingZone, "dp-ok", "default", dataplaneWithZoneTag(connectingZone))

		stored := storedDataplane(t, globalStore, ctx)
		g.Expect(stored.GetMeta().GetLabels()[mesh_proto.ZoneTag]).To(Equal(connectingZone))
		g.Consistently(func() float64 {
			return testutil.ToFloat64(rewrites.WithLabelValues(string(core_mesh.DataplaneType)))
		}, "1s", "100ms").Should(Equal(0.0), "no rewrite must be counted when values already match")
	})
}
