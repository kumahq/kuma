// Package perf contains a local, in-process load harness for the Global CP KDS
// delta server. It is skipped unless KUMA_KDS_PERF=1 so it never runs in CI.
//
// Usage:
//
//	KUMA_KDS_PERF=1 KDS_ZONES=50 KDS_RESYNC=1s go test ./pkg/test/kds/perf/ \
//	    -run TestGlobalKDSScale -cpuprofile cpu.out -memprofile mem.out -timeout 10m
package perf

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	config_types "github.com/kumahq/kuma/v3/pkg/config/types"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	zone_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zone/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/events"
	kds_client "github.com/kumahq/kuma/v3/pkg/kds/client"
	kds_context "github.com/kumahq/kuma/v3/pkg/kds/context"
	"github.com/kumahq/kuma/v3/pkg/kds/mux"
	kds_server "github.com/kumahq/kuma/v3/pkg/kds/server"
	"github.com/kumahq/kuma/v3/pkg/multitenant"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	test_grpc "github.com/kumahq/kuma/v3/pkg/test/grpc"
	"github.com/kumahq/kuma/v3/pkg/test/kds/setup"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

// countingStore counts store operations so a run can be expressed in store
// reads per second, which is what the Global CP actually pays for at scale.
type countingStore struct {
	store.ResourceStore

	gets    atomic.Int64
	lists   atomic.Int64
	creates atomic.Int64
	updates atomic.Int64
	deletes atomic.Int64

	listsByType sync.Map
	getsByType  sync.Map
}

func (c *countingStore) bump(m *sync.Map, t core_model.ResourceType) {
	v, _ := m.LoadOrStore(t, &atomic.Int64{})
	v.(*atomic.Int64).Add(1)
}

func (c *countingStore) Get(ctx context.Context, r core_model.Resource, fs ...store.GetOptionsFunc) error {
	c.gets.Add(1)
	c.bump(&c.getsByType, r.Descriptor().Name)
	return c.ResourceStore.Get(ctx, r, fs...)
}

func (c *countingStore) List(ctx context.Context, l core_model.ResourceList, fs ...store.ListOptionsFunc) error {
	c.lists.Add(1)
	c.bump(&c.listsByType, l.GetItemType())
	return c.ResourceStore.List(ctx, l, fs...)
}

func (c *countingStore) Create(ctx context.Context, r core_model.Resource, fs ...store.CreateOptionsFunc) error {
	c.creates.Add(1)
	return c.ResourceStore.Create(ctx, r, fs...)
}

func (c *countingStore) Update(ctx context.Context, r core_model.Resource, fs ...store.UpdateOptionsFunc) error {
	c.updates.Add(1)
	return c.ResourceStore.Update(ctx, r, fs...)
}

func (c *countingStore) Delete(ctx context.Context, r core_model.Resource, fs ...store.DeleteOptionsFunc) error {
	c.deletes.Add(1)
	return c.ResourceStore.Delete(ctx, r, fs...)
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envDur(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func zoneName(i int) string { return fmt.Sprintf("zone-%d", i) }

func meshName(i int) string { return fmt.Sprintf("mesh-%d", i) }

// seed populates the global store with a workload shaped like a real tenant: a
// Zone per connected zone, a set of meshes, and policies across every type the
// Global CP ships down to zones.
func seed(t *testing.T, st store.ResourceStore, meshes, zones, policiesPerType int, types []core_model.ResourceType) {
	t.Helper()
	ctx := context.Background()
	for z := range zones {
		zone := zone_api.NewZoneResource()
		zone.Spec = &zone_api.Zone{Enabled: pointer.To(true)}
		if err := st.Create(ctx, zone, store.CreateByKey(zoneName(z), core_model.NoMesh)); err != nil {
			t.Fatalf("seed zone: %v", err)
		}
	}
	for m := range meshes {
		mr := core_mesh.NewMeshResource()
		mr.Spec = &mesh_proto.Mesh{}
		if err := st.Create(ctx, mr, store.CreateByKey(meshName(m), core_model.NoMesh)); err != nil {
			t.Fatalf("seed mesh: %v", err)
		}
	}

	created := 0
	zoneOriginated := 0
	for _, typ := range types {
		if typ == core_mesh.MeshType || typ == zone_api.ZoneType {
			continue
		}
		desc, err := registry.Global().DescriptorFor(typ)
		if err != nil {
			continue
		}
		for m := range meshes {
			for p := range policiesPerType {
				res, err := registry.Global().NewObject(typ)
				if err != nil {
					continue
				}
				meshOf := core_model.NoMesh
				if desc.Scope == core_model.ScopeMesh {
					meshOf = meshName(m)
				}
				key := fmt.Sprintf("%s-%d-%d", typ, m, p)
				opts := []store.CreateOptionsFunc{store.CreateByKey(key, meshOf)}
				if desc.KDSFlags.Has(core_model.SyncedAcrossZonesFlag) {
					opts = append(opts, store.CreateWithLabels(map[string]string{
						mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
						mesh_proto.ZoneTag:             zoneName(p % zones),
					}))
					zoneOriginated++
				}
				if err := st.Create(ctx, res, opts...); err == nil {
					created++
				}
			}
			if desc.Scope != core_model.ScopeMesh {
				break
			}
		}
	}
	t.Logf("seeded %d zones, %d meshes, %d resources across %d types (%d zone-originated)", zones, meshes, created, len(types), zoneOriginated)
}

func runChurn(ctx context.Context, st store.ResourceStore, perSec int) {
	ticker := time.NewTicker(time.Second / time.Duration(perSec))
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			res := core_mesh.NewMeshResource()
			if err := st.Get(ctx, res, store.GetByKey(meshName(0), core_model.NoMesh)); err != nil {
				i++
				continue
			}
			_ = st.Update(ctx, res, store.UpdateWithLabels(map[string]string{"churn": strconv.Itoa(i)}))
			i++
		}
	}
}

func TestGlobalKDSScale(t *testing.T) {
	if os.Getenv("KUMA_KDS_PERF") != "1" {
		t.Skip("set KUMA_KDS_PERF=1 to run the KDS load harness")
	}

	cfgDefaults := kuma_cp.DefaultConfig()
	zones := envInt("KDS_ZONES", 50)
	meshes := envInt("KDS_MESHES", 5)
	perType := envInt("KDS_POLICIES_PER_TYPE", 10)
	churn := envInt("KDS_CHURN_PER_SEC", 0)
	cacheTTL := envDur("KDS_STORE_CACHE_TTL", 0)
	stagger := envDur("KDS_CONNECT_STAGGER", 0)
	duration := envDur("KDS_DURATION", 30*time.Second)
	flush := envDur("KDS_FLUSH", cfgDefaults.Multizone.Global.KDS.EventBasedWatchdog.FlushInterval.Duration)
	resync := envDur("KDS_RESYNC", cfgDefaults.Multizone.Global.KDS.EventBasedWatchdog.FullResyncInterval.Duration)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := kuma_cp.DefaultConfig()
	cfg.Mode = config_core.Global
	cfg.Multizone.Global.KDS.EventBasedWatchdog.FlushInterval = config_types.Duration{Duration: flush}
	cfg.Multizone.Global.KDS.EventBasedWatchdog.FullResyncInterval = config_types.Duration{Duration: resync}

	memStore := memory.NewStore()
	cs := &countingStore{ResourceStore: memStore}
	rt := setup.NewTestRuntime(ctx, cfg, cs)
	memStore.(interface{ SetEventWriter(events.Emitter) }).SetEventWriter(rt.EventBus())
	if cacheTTL > 0 {
		cached, err := manager.NewCachedManager(rt.ReadOnlyResourceManager(), cacheTTL, rt.Metrics(), multitenant.SingleTenant)
		if err != nil {
			t.Fatalf("cached manager: %v", err)
		}
		rt.SetReadOnlyResourceManager(cached)
	}

	kdsCtx := kds_context.DefaultContext(ctx, rt.ReadOnlyResourceManager(), cfg)
	types := kdsCtx.TypesSentByGlobal

	seed(t, cs, meshes, zones, perType, types)

	srv, _, err := kds_server.New(
		core.Log.WithName("kds-perf"),
		rt,
		types,
		"global",
		kdsCtx.GlobalProvidedFilter,
		kdsCtx.GlobalResourceMapper,
		time.Second,
		cfg.Multizone.Global.KDS.EventBasedWatchdog.AsRuntimeConfig(),
	)
	if err != nil {
		t.Fatalf("kds server: %v", err)
	}

	cs.gets.Store(0)
	cs.lists.Store(0)
	cs.listsByType.Range(func(_, v any) bool { v.(*atomic.Int64).Store(0); return true })
	cs.getsByType.Range(func(_, v any) bool { v.(*atomic.Int64).Store(0); return true })

	stop := make(chan struct{})
	var wg sync.WaitGroup
	streams := make([]*test_grpc.MockDeltaClientStream, 0, zones)
	names := make([]string, 0, zones)
	var responses atomic.Int64
	var nacks atomic.Int64

	for z := range zones {
		serverStream := test_grpc.NewMockDeltaServerStream()
		wg.Go(func() {
			_ = srv.DeltaStreamHandler(mux.NewErrorRecorderStream(serverStream), "")
		})
		streams = append(streams, serverStream.ClientStream(stop))
		names = append(names, zoneName(z))
	}

	cb := &kds_client.Callbacks{
		OnResourcesReceived: func(_ kds_client.UpstreamResponse) (error, error) {
			responses.Add(1)
			return nil, nil
		},
		OnNACK: func(_ core_model.ResourceType) {
			nacks.Add(1)
		},
	}
	if stagger > 0 {
		gap := stagger / time.Duration(zones)
		for i := range streams {
			setup.StartDeltaClient(streams[i:i+1], names[i:i+1], types, stop, cb)
			time.Sleep(gap)
		}
	} else {
		setup.StartDeltaClient(streams, names, types, stop, cb)
	}

	if churn > 0 {
		go runChurn(ctx, cs, churn)
	}

	start := time.Now()
	time.Sleep(duration)
	elapsed := time.Since(start).Seconds()

	lists := cs.lists.Load()
	gets := cs.gets.Load()

	t.Logf("=== zones=%d meshes=%d types=%d flush=%s resync=%s cacheTTL=%s over %.0fs ===",
		zones, meshes, len(types), flush, resync, cacheTTL, elapsed)
	t.Logf("store LIST : %7d total %9.1f/s %7.2f/s per zone", lists, float64(lists)/elapsed, float64(lists)/elapsed/float64(zones))
	t.Logf("store GET  : %7d total %9.1f/s %7.2f/s per zone", gets, float64(gets)/elapsed, float64(gets)/elapsed/float64(zones))
	t.Logf("KDS responses delivered: %d (%.1f/s), client NACKs: %d", responses.Load(), float64(responses.Load())/elapsed, nacks.Load())

	cs.getsByType.Range(func(k, v any) bool {
		if n := v.(*atomic.Int64).Load(); n > 0 {
			t.Logf("  GET %-32s %8d (%.1f/s)", string(k.(core_model.ResourceType)), n, float64(n)/elapsed)
		}
		return true
	})

	close(stop)
	cancel()
}
