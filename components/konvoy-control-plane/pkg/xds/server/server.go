package server

import (
	"context"
	"log"
	"sync"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util"
)

var (
	wg sync.WaitGroup
)

type RunArgs struct {
	GrpcPort        int
	HttpPort        int
	DiagnosticsPort int
}

func Run(args RunArgs) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		wg.Wait()
	}()

	nodes := make(chan *v2_core.Node, 10)
	hasher := &hasher{}
	logger := logger{}
	configs := cache.NewSnapshotCache(true, hasher, logger)
	cb := &callbacks{nodes}
	srv := server.NewServer(configs, cb)

	// start xDS server
	go func(wg sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		RunGrpcServer(ctx, srv, args.GrpcPort)
	}(wg)
	go func(wg sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		RunHttpGateway(ctx, srv, args.HttpPort)
	}(wg)

	// start reconciliation loop
	go func(wg sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		RunSyncer(ctx, configs, hasher, nodes)
	}(wg)

	// start reconciliation loop
	go func(wg sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		RunDiagnosticsServer(ctx, args.DiagnosticsPort)
	}(wg)

	util.WaitStopSignal()
	return nil
}

type hasher struct {
}

func (h *hasher) ID(node *v2_core.Node) string {
	if node == nil {
		return "unknown"
	}
	return node.Id
}

type logger struct{
}

func (logger logger) Infof(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}
func (logger logger) Errorf(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}

type callbacks struct {
	nodes   chan<- *v2_core.Node
}

func (cb *callbacks) OnStreamOpen(_ context.Context, id int64, typ string) error {
	return nil
}
func (cb *callbacks) OnStreamClosed(id int64) {
}
func (cb *callbacks) OnStreamRequest(_ int64, req *v2.DiscoveryRequest) error {
	if req.Node != nil {
		cb.nodes <- req.Node
	}
	return nil
}
func (cb *callbacks) OnStreamResponse(int64, *v2.DiscoveryRequest, *v2.DiscoveryResponse) {}
func (cb *callbacks) OnFetchRequest(_ context.Context, req *v2.DiscoveryRequest) error {
	if req.Node != nil {
		cb.nodes <- req.Node
	}
	return nil
}
func (cb *callbacks) OnFetchResponse(*v2.DiscoveryRequest, *v2.DiscoveryResponse) {}
