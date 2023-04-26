package components

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/kumahq/kuma/pkg/core/runtime"
	metrics "github.com/kumahq/kuma/pkg/metrics/store"
	"github.com/kumahq/kuma/pkg/version"
)

func Setup(rt runtime.Runtime) error {
	cpInfoMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
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
		},
	}, func() float64 {
		return 1.0
	})
	if err := rt.Metrics().Register(cpInfoMetric); err != nil {
		return err
	}

	leaderMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "leader",
		Help: "1 indicates that this instance is leader",
	}, func() float64 {
		if rt.LeaderInfo().IsLeader() {
			return 1.0
		} else {
			return 0.0
		}
	})
	if err := rt.Metrics().Register(leaderMetric); err != nil {
		return err
	}

	if err := rt.Metrics().Register(collectors.NewGoCollector()); err != nil {
		return err
	}
	if err := rt.Metrics().Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		return err
	}

	// We don't want to use cached ResourceManager because the cache is just for a couple of seconds
	// and we will be retrieving resources every minute. There is no other place in the system for now that needs all resources from all meshes
	// therefore it makes no sense to cache all content of the Database in the cache.
	counter, err := metrics.NewStoreCounter(rt.ResourceManager(), rt.Metrics())
	if err != nil {
		return err
	}
	if err := rt.Add(counter); err != nil {
		return err
	}
	return nil
}
