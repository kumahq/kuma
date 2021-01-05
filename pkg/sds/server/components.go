package server

import (
	"google.golang.org/grpc"

	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/sds/server/metrics"
	v2 "github.com/kumahq/kuma/pkg/sds/server/v2"
)

func RegisterSDS(rt core_runtime.Runtime, server *grpc.Server) error {
	sdsMetrics, err := metrics.NewSDSMetrics(rt.Metrics())
	if err != nil {
		return err
	}

	if err := v2.RegisterSDS(rt, sdsMetrics, server); err != nil {
		return err
	}
	return nil
}
