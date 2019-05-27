package server

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
)

var (
	source = NewDummySource()
)

func RunSyncer(ctx context.Context, config cache.SnapshotCache, hasher cache.NodeHash, nodes <-chan *core.Node) {
	for {
		select {
		case node, open := <-nodes:
			if !open {
				return
			}

			snapshot := source.NewSnapshot(node)
			if err := snapshot.Consistent(); err != nil {
				log.Printf("snapshot inconsistency: %+v\n", snapshot)
			}

			err := config.SetSnapshot(hasher.ID(node), snapshot)
			if err != nil {
				log.Printf("snapshot error %q for %+v\n", err, snapshot)
				os.Exit(1)
			}
		case <-ctx.Done():
			return
		}
	}
}

type SnapshotSource interface {
	NewSnapshot(node *core.Node) cache.Snapshot
}

func NewDummySource() SnapshotSource {
	return &dummySource{}
}

type dummySource struct {
}

func (s *dummySource) NewSnapshot(node *core.Node) cache.Snapshot {
	clusters := make([]cache.Resource, 2)
	endpoints := make([]cache.Resource, 2)
	for i := 0; i < 2; i++ {
		port := uint32(8080 + i)
		name := fmt.Sprintf("localhost:%d", port)
		clusters[i] = CreateCluster(name)
		endpoints[i] = CreateEndpoint(name, port)
	}

	listeners := make([]cache.Resource, 2)
	listeners[0] = CreateInboundListener("inbound-listener-18080", 18080, "localhost:8080")
	listeners[1] = CreateInboundListener("inbound-listener-18081", 18081, "localhost:8081")

	version := "v1"
	out := cache.Snapshot{
		Endpoints: cache.NewResources(version, endpoints),
		Clusters:  cache.NewResources(version, clusters),
		Routes:    cache.NewResources(version, []cache.Resource{}),
		Listeners: cache.NewResources(version, listeners),
	}

	return out
}
