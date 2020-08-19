package components

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/version"
)

func Setup(rt runtime.Runtime) {
	grpc_prometheus.EnableHandlingTimeHistogram() // setup histograms for all gRPC metrics

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "cp_info",
		Help: "Static information about the CP instance",
		ConstLabels: map[string]string{
			"instance_id": rt.GetInstanceId(),
			"cluster_id":  rt.GetClusterId(),
			"product":     version.Product,
			"version":     version.Build.Version,
			"build_date":  version.Build.BuildDate,
			"git_commit":  version.Build.GitCommit,
			"git_tag":     version.Build.GitTag,
			"mode":        rt.Config().Mode,
		},
	}, func() float64 {
		return 1.0
	})

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "leader",
		Help: "1 indicates that this instance is leader",
	}, func() float64 {
		if rt.LeaderInfo().IsLeader() {
			return 1.0
		} else {
			return 0.0
		}
	})
}
