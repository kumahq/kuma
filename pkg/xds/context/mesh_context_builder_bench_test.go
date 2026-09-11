package context_test

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_apis "github.com/kumahq/kuma/v3/pkg/core/resources/apis"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/multitenant"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	test_store "github.com/kumahq/kuma/v3/pkg/test/store"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	xds_server "github.com/kumahq/kuma/v3/pkg/xds/server"
)

type meshSize struct {
	services   int
	dataplanes int
	policies   int
}

func (s meshSize) String() string {
	return fmt.Sprintf("svc=%d/dp=%d/pol=%d", s.services, s.dataplanes, s.policies)
}

var benchSizes = []meshSize{
	{services: 50, dataplanes: 500, policies: 100},
	{services: 200, dataplanes: 2000, policies: 500},
}

// BenchmarkMeshContextUnchangedCachedStore is BenchmarkMeshContextUnchanged behind
// the cached manager, as within the store cache TTL in the control plane, so it
// shows the builder's own cost without the store decoding resources.
func BenchmarkMeshContextUnchangedCachedStore(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(size.String(), func(b *testing.B) {
			s := memory.NewStore()
			seedMesh(b, s, size)
			metrics, err := core_metrics.NewMetrics("Zone")
			if err != nil {
				b.Fatal(err)
			}
			cached, err := manager.NewCachedManager(s, time.Hour, metrics, multitenant.SingleTenant)
			if err != nil {
				b.Fatal(err)
			}
			builder := newBenchBuilder(cached)
			latest, err := builder.BuildIfChanged(context.Background(), "mesh-1", nil)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				latest, err = builder.BuildIfChanged(context.Background(), "mesh-1", latest)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func dataplaneYAML(i, services int, address string) string {
	return fmt.Sprintf(`
type: Dataplane
name: dp-%d
mesh: mesh-1
labels:
  app: svc-%d
networking:
  address: %s
  inbound:
  - port: 8080
`, i, i%services, address)
}

func seedMesh(b *testing.B, s store.ResourceStore, size meshSize) {
	b.Helper()
	plugins.InitAll(core_apis.NameToModule)
	plugins.InitAll(policies.NameToModule)
	var sb strings.Builder
	sb.WriteString("type: Mesh\nname: mesh-1\n")
	for i := range size.services {
		fmt.Fprintf(&sb, `---
type: MeshService
name: svc-%d
mesh: mesh-1
spec:
  selector:
    dataplaneTags:
      app: svc-%d
  ports:
  - port: 80
    targetPort: 8080
    appProtocol: http
status:
  vips:
  - ip: 240.0.%d.%d
`, i, i, i/250, i%250)
	}
	for i := range size.dataplanes {
		sb.WriteString("---")
		sb.WriteString(dataplaneYAML(i, size.services, fmt.Sprintf("10.%d.%d.%d", i/65025, (i/255)%255, i%255+1)))
	}
	for i := range size.policies {
		fmt.Fprintf(&sb, `---
type: MeshTimeout
name: mt-%d
mesh: mesh-1
spec:
  to:
  - targetRef:
      kind: Mesh
    default:
      connectionTimeout: %ds
`, i, i+1)
	}
	if err := test_store.LoadResources(context.Background(), s, sb.String()); err != nil {
		b.Fatal(err)
	}
}

func newBenchBuilder(rm manager.ReadOnlyResourceManager) xds_context.MeshContextBuilder {
	return xds_context.NewMeshContextBuilder(
		rm,
		xds_server.MeshResourceTypes(),
		func(host string) ([]net.IP, error) { return []net.IP{net.ParseIP(host)}, nil },
		"zone-1",
		xds_context.WithPolicyMatchingHash(),
	)
}

// BenchmarkMeshContextUnchanged measures the steady-state cost paid every mesh
// cache expiry (1s by default) when nothing in the mesh changed.
func BenchmarkMeshContextUnchanged(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(size.String(), func(b *testing.B) {
			s := memory.NewStore()
			seedMesh(b, s, size)
			builder := newBenchBuilder(s)
			latest, err := builder.BuildIfChanged(context.Background(), "mesh-1", nil)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				latest, err = builder.BuildIfChanged(context.Background(), "mesh-1", latest)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMeshContextDataplaneChanged measures a rebuild triggered by a single
// Dataplane changing, the most common change during rollouts.
func BenchmarkMeshContextDataplaneChanged(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(size.String(), func(b *testing.B) {
			s := memory.NewStore()
			seedMesh(b, s, size)
			builder := newBenchBuilder(s)
			latest, err := builder.BuildIfChanged(context.Background(), "mesh-1", nil)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			i := 0
			for b.Loop() {
				b.StopTimer()
				i++
				if err := test_store.LoadResources(context.Background(), s, dataplaneYAML(0, size.services, fmt.Sprintf("10.255.%d.%d", (i/250)%250, i%250+1))); err != nil {
					b.Fatal(err)
				}
				b.StartTimer()
				latest, err = builder.BuildIfChanged(context.Background(), "mesh-1", latest)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMeshContextPolicyChanged measures a rebuild triggered by a single
// policy changing.
func BenchmarkMeshContextPolicyChanged(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(size.String(), func(b *testing.B) {
			s := memory.NewStore()
			seedMesh(b, s, size)
			builder := newBenchBuilder(s)
			latest, err := builder.BuildIfChanged(context.Background(), "mesh-1", nil)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			i := 0
			for b.Loop() {
				b.StopTimer()
				i++
				if err := test_store.LoadResources(context.Background(), s, fmt.Sprintf(`
type: MeshTimeout
name: mt-0
mesh: mesh-1
spec:
  to:
  - targetRef:
      kind: Mesh
    default:
      connectionTimeout: %ds
`, 1000+i)); err != nil {
					b.Fatal(err)
				}
				b.StartTimer()
				latest, err = builder.BuildIfChanged(context.Background(), "mesh-1", latest)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
