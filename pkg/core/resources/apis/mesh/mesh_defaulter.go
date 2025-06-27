package mesh

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/proto"
)

func (mesh *MeshResource) Default() error {
	// default settings for Prometheus metrics
	for idx, backend := range mesh.Spec.GetMetrics().GetBackends() {
		if backend.GetType() != mesh_proto.MetricsPrometheusType {
			continue
		}
		cfg := mesh_proto.PrometheusMetricsBackendConfig{}
		if err := proto.ToTyped(backend.GetConf(), &cfg); err != nil {
			return errors.Wrap(err, "could not convert the backend")
		}

		if cfg.SkipMTLS == nil && cfg.Tls == nil {
			cfg.Tls = &mesh_proto.PrometheusTlsConfig{
				Mode: mesh_proto.PrometheusTlsConfig_activeMTLSBackend,
			}
		}
		if cfg.Tls == nil && cfg.SkipMTLS != nil && cfg.SkipMTLS.Value {
			cfg.Tls = &mesh_proto.PrometheusTlsConfig{
				Mode: mesh_proto.PrometheusTlsConfig_disabled,
			}
		}
		if cfg.Tls == nil && cfg.SkipMTLS != nil && !cfg.SkipMTLS.Value {
			cfg.Tls = &mesh_proto.PrometheusTlsConfig{
				Mode: mesh_proto.PrometheusTlsConfig_activeMTLSBackend,
			}
		}
		if cfg.Tls != nil && cfg.SkipMTLS != nil && cfg.Tls.GetMode() == mesh_proto.PrometheusTlsConfig_providedTLS {
			cfg.Tls = &mesh_proto.PrometheusTlsConfig{
				Mode: mesh_proto.PrometheusTlsConfig_providedTLS,
			}
			cfg.SkipMTLS = nil
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
			return errors.Wrap(err, "could not convert the backend")
		}
		mesh.Spec.Metrics.Backends[idx].Conf = str
	}
	return nil
}
