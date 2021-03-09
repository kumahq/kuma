package server

import (
	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/mads"
	mads_v1 "github.com/kumahq/kuma/pkg/mads/v1/service"
	mads_v1alpha1 "github.com/kumahq/kuma/pkg/mads/v1alpha1/service"
)

var (
	madsServerLog = core.Log.WithName("mads-server")
)

func SetupServer(rt core_runtime.Runtime) error {
	config := rt.Config().MonitoringAssignmentServer
	rm := rt.ReadOnlyResourceManager()

	var grpcServices []GrpcService
	var httpServices []HttpService

	if config.VersionIsEnabled(mads.API_V1_ALPHA1) {
		madsServerLog.Info("MADS v1alpha1 is enabled")
		svc := mads_v1alpha1.NewService(config, rm, madsServerLog.WithValues("apiVersion", mads.API_V1_ALPHA1))
		grpcServices = append(grpcServices, svc)
	}

	if config.VersionIsEnabled(mads.API_V1) {
		madsServerLog.Info("MADS v1 is enabled")
		svc := mads_v1.NewService(config, rm, madsServerLog.WithValues("apiVersion", mads.API_V1))
		grpcServices = append(grpcServices, svc)
		httpServices = append(httpServices, svc)
	}

	if config.GrpcEnabled {
		if err := rt.Add(&grpcServer{
			services: grpcServices,
			config: config,
			metrics: rt.Metrics(),
		}); err != nil {
			return err
		}
	}

	if config.HttpEnabled {
		if err := rt.Add(&httpServer{
			services: httpServices,
			config: config,
			metrics: rt.Metrics(),
		}); err != nil {
			return err
		}
	}

	return nil
}
