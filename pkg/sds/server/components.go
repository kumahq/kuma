package server

import (
	"google.golang.org/grpc"

	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	sds_metrics "github.com/kumahq/kuma/pkg/sds/metrics"
	v2 "github.com/kumahq/kuma/pkg/sds/server/v2"
)

func RegisterSDS(rt core_runtime.Runtime, server *grpc.Server) error {
	sdsMetrics, err := sds_metrics.NewMetrics(rt.Metrics())
	if err != nil {
		return err
	}

	if err := v2.RegisterSDS(rt, sdsMetrics, server); err != nil {
		return err
	}
	return nil
}
