package components

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/kumahq/kuma/pkg/core/runtime"
	metrics "github.com/kumahq/kuma/pkg/metrics/store"
	"github.com/kumahq/kuma/pkg/version"
)

func Setup(rt runtime.Runtime) error {
	labels := version.Build.AsMap()
	labels["instance_id"] = rt.GetInstanceId()
	labels["cluster_id"] = rt.GetClusterId()
	cpInfoMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "cp_info",
		Help:        "Static information about the CP instance",
		ConstLabels: labels,
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
		if !isDuplicatedError(err) { // can be already registered when K8S registerer is used
			return err
		}
	}
	if err := rt.Metrics().Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		if !isDuplicatedError(err) { // can be already registered when K8S registerer is used
			return err
		}
	}

	// We don't want to use cached ResourceManager because the cache is just for a couple of seconds
	// and we will be retrieving resources every minute. There is no other place in the system for now that needs all resources from all meshes
	// therefore it makes no sense to cache all content of the Database in the cache.
	if rt.Config().Metrics.ControlPlane.ReportResourcesCount {
		counter, err := metrics.NewStoreCounter(rt.ResourceManager(), rt.Metrics(), rt.Tenants())
		if err != nil {
			return err
		}
		if err := rt.Add(counter); err != nil {
			return err
		}
	}

	return nil
}

func isDuplicatedError(err error) bool {
	return strings.HasPrefix(err.Error(), "a previously registered descriptor with the same fully-qualified name as") &&
		strings.HasSuffix(err.Error(), "has different label names or a different help string")
}
