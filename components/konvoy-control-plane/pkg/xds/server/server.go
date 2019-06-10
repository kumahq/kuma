package server

import (
	"fmt"

	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/model/api/v1alpha1"
	model_controllers "github.com/Kong/konvoy/components/konvoy-control-plane/model/controllers"
	util_manager "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/manager"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
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
	store := cache.NewSnapshotCache(true, hasher, logger)
	reconciler := reconciler{&templateSnapshotGenerator{
		Template: &konvoy_mesh.ProxyTemplate{
			Spec: konvoy_mesh.ProxyTemplateSpec{
				Sources: []konvoy_mesh.ProxyTemplateSource{
					{
						Profile: &konvoy_mesh.ProxyTemplateProfileSource{
							Name: "transparent-inbound-proxy",
						},
					},
					{
						Profile: &konvoy_mesh.ProxyTemplateProfileSource{
							Name: "transparent-outbound-proxy",
						},
					},
				},
			},
		},
	}, &simpleSnapshotCacher{hasher, store}}
	srv := xds.NewServer(store, nil)

	if err := util_manager.SetupWithManager(
		mgr,
		&model_controllers.ProxyTemplateReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("ProxyTemplate"),
		},
		&model_controllers.ProxyReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("Proxy"),
		},
		&model_controllers.PodReconciler{
			Client:   mgr.GetClient(),
			Log:      ctrl.Log.WithName("controllers").WithName("Proxy"),
			Observer: &reconciler,
		},
	); err != nil {
		return err
	}

	return util_manager.Add(mgr,
		// xDS gRPC API
		&grpcServer{srv, s.Args.GrpcPort},
		// xDS HTTP API
		&httpGateway{srv, s.Args.HttpPort},
		// diagnostics server
		&diagnosticsServer{s.Args.DiagnosticsPort})
}

type hasher struct {
}

func (h hasher) ID(node *envoy_core.Node) string {
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
