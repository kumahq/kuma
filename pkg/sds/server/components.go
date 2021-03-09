package server

import (
	"github.com/pkg/errors"

	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	sds_metrics "github.com/kumahq/kuma/pkg/sds/metrics"
	v2 "github.com/kumahq/kuma/pkg/sds/server/v2"
	v3 "github.com/kumahq/kuma/pkg/sds/server/v3"
)

func Setup(rt core_runtime.Runtime) error {
	sdsMetrics, err := sds_metrics.NewMetrics(rt.Metrics())
	if err != nil {
		return err
	}

	if err := v2.RegisterSDS(rt, sdsMetrics); err != nil {
		return errors.Wrap(err, "could not register V2 SDS")
	}
	if err := v3.RegisterSDS(rt, sdsMetrics); err != nil {
		return errors.Wrap(err, "could not register V3 SDS")
	}
	return nil
}
