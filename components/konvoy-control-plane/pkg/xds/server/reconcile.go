package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	reconcileLog = ctrl.Log.WithName("xds-server").WithName("reconcile")
)

type reconciler struct {
	nodes     <-chan *core.Node
	generator snapshotGenerator
	hasher    cache.NodeHash
	store     cache.SnapshotCache
}

// Make sure that reconciler implements all relevant interfaces
var (
	_ manager.Runnable               = &reconciler{}
	_ manager.LeaderElectionRunnable = &reconciler{}
)

func (r *reconciler) Start(stop <-chan struct{}) error {
	for {
		select {
		case node, open := <-r.nodes:
			if !open {
				return nil
			}

			snapshot := r.generator.NewSnapshot(node)
			if err := snapshot.Consistent(); err != nil {
				reconcileLog.Error(err, "inconsistent snapshot", "snapshot", snapshot)
			}

			err := r.store.SetSnapshot(r.hasher.ID(node), snapshot)
			if err != nil {
				reconcileLog.Error(err, "failed to store snapshot", "snapshot", snapshot)
				return err
			}
		case <-stop:
			return nil
		}
	}
}

func (r *reconciler) NeedLeaderElection() bool {
	return false
}

type snapshotGenerator interface {
	NewSnapshot(node *core.Node) cache.Snapshot
}

type basicSnapshotGenerator struct {
}

func (s *basicSnapshotGenerator) NewSnapshot(node *core.Node) cache.Snapshot {
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
	ports     []uint32
}

func parseNodeInfo(node *core.Node) nodeInfo {
	nodeAddresses := "127.0.0.1"
	nodePorts := ""
	if node.Metadata != nil && node.Metadata.Fields != nil {
		if value := node.Metadata.Fields["IPS"]; value != nil && value.GetStringValue() != "" {
			nodeAddresses = value.GetStringValue()
		}
		if value := node.Metadata.Fields["PORTS"]; value != nil && value.GetStringValue() != "" {
			nodePorts = value.GetStringValue()
		}
	}

	addresses := make([]string, 0, 1)
	for _, nodeAddress := range strings.Split(nodeAddresses, ",") {
		address := strings.TrimSpace(nodeAddress)
		if address == "" {
			continue
		}
		addresses = append(addresses, address)
	}

	ports := make([]uint32, 0, 1)
	for _, nodePort := range strings.Split(nodePorts, ",") {
		port, err := strconv.ParseUint(strings.TrimSpace(nodePort), 10, 32)
		if err != nil {
			continue
		}
		ports = append(ports, uint32(port))
	}

	return nodeInfo{addresses, ports}
}
