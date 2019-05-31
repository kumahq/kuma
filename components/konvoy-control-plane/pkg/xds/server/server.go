package server

import (
	"context"
	"fmt"

	util_manager "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/manager"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	xdsServerLog = ctrl.Log.WithName("xds-server")
)

type RunArgs struct {
	GrpcPort        int
	HttpPort        int
	DiagnosticsPort int
}

type Server struct {
	Args RunArgs
}

func (s *Server) SetupWithManager(mgr ctrl.Manager) error {
	hasher := hasher{}
	logger := logger{}
	nodes := make(chan *v2_core.Node, 10)
	store := cache.NewSnapshotCache(true, hasher, logger)
	cb := &callbacks{nodes}
	srv := xds.NewServer(store, cb)

	return util_manager.Add(mgr,
		// xDS gRPC API
		&grpcServer{srv, s.Args.GrpcPort},
		// xDS HTTP API
		&httpGateway{srv, s.Args.HttpPort},
		// reconciliation loop
		&reconciler{nodes, &basicSnapshotGenerator{}, hasher, store},
		// diagnostics server
		&diagnosticsServer{s.Args.DiagnosticsPort})
}

type hasher struct {
}

func (h hasher) ID(node *v2_core.Node) string {
	if node == nil {
		return "unknown"
	}
	return node.Id
}

type logger struct {
}

func (logger logger) Infof(format string, args ...interface{}) {
	xdsServerLog.V(1).Info(fmt.Sprintf(format, args...))
}
func (logger logger) Errorf(format string, args ...interface{}) {
	xdsServerLog.Error(fmt.Errorf(format, args...), "")
}

type callbacks struct {
	nodes chan<- *v2_core.Node
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
