package mesh

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/proto"
)

func (mesh *MeshResource) Default() error {
	// default settings for Prometheus metrics
	for idx, backend := range mesh.Spec.GetMetrics().GetBackends() {
		if backend.GetType() == mesh_proto.MetricsPrometheusType {
			cfg := mesh_proto.PrometheusMetricsBackendConfig{}
			if err := proto.ToTyped(backend.GetConf(), &cfg); err != nil {
				return fmt.Errorf("could not convert the backend: %w", err)
			}

			if cfg.Port == 0 {
				cfg.Port = 5670
			}
			if cfg.Path == "" {
				cfg.Path = "/metrics"
			}
			if len(cfg.Tags) == 0 {
				cfg.Tags = map[string]string{
					mesh_proto.ServiceTag: "dataplane-metrics",
				}
			}

			str, err := proto.ToStruct(&cfg)
			if err != nil {
				return fmt.Errorf("could not convert the backend: %w", err)
			}
			mesh.Spec.Metrics.Backends[idx].Conf = str
		}
	}
	return nil
}
