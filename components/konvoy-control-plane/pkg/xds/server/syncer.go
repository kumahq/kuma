package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"strconv"

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
	nodeInfo := parseNodeInfo(node)

	listeners := make([]cache.Resource, 0, 2)
	listeners = append(listeners, CreateCatchAllListener("catch_all", "0.0.0.0", 15001, "pass_through"))

	clusters := make([]cache.Resource, 0, 2)
	clusters = append(clusters, CreatePassThroughCluster("pass_through"))

	for _, port := range nodeInfo.ports {
		localClusterName := fmt.Sprintf("localhost:%d", port)
		clusters = append(clusters, CreateLocalCluster(localClusterName, "127.0.0.1", port))

		for _, address := range nodeInfo.addresses {
			inboundListenerName := fmt.Sprintf("inbound:%s:%d", address, port)

			listeners = append(listeners, CreateInboundListener(inboundListenerName, address, port, localClusterName))
		}
	}

	version := "v1"
	out := cache.Snapshot{
		Endpoints: cache.NewResources(version, []cache.Resource{}),
		Clusters:  cache.NewResources(version, clusters),
		Routes:    cache.NewResources(version, []cache.Resource{}),
		Listeners: cache.NewResources(version, listeners),
		Secrets:   cache.NewResources(version, []cache.Resource{}),
	}

	return out
}

type nodeInfo struct {
	addresses []string
	ports []uint32
}

func parseNodeInfo(node *core.Node) nodeInfo {
	node_addresses := "127.0.0.1"
	node_ports := ""
	if node.Metadata != nil && node.Metadata.Fields != nil {
		if value := node.Metadata.Fields["IPS"]; value != nil && value.GetStringValue() != "" {
			node_addresses = value.GetStringValue()
		}
		if value := node.Metadata.Fields["PORTS"]; value != nil && value.GetStringValue() != "" {
			node_ports = value.GetStringValue()
		}
	}

	addresses := make([]string, 0, 1)
	for _, node_address := range strings.Split(node_addresses, ",") {
		address := strings.TrimSpace(node_address)
		if address == "" {
			continue
		}
		addresses = append(addresses, address)
	}

	ports := make([]uint32, 0, 1)
	for _, node_port := range strings.Split(node_ports, ",") {
		port, err := strconv.ParseUint(strings.TrimSpace(node_port), 10, 32)
		if err != nil {
			continue
		}
		ports = append(ports, uint32(port))
	}

	return nodeInfo{addresses, ports}
}
