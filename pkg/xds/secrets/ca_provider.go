package secrets

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

type CaProvider interface {
	// Get returns all PEM encoded CAs, a list of CAs that were used to generate a secret and an error.
	Get(context.Context, *core_mesh.MeshResource) (*core_xds.CaSecret, []string, error)
}

func NewCaProvider(caManagers core_ca.Managers, metrics core_metrics.Metrics) (CaProvider, error) {
	latencyMetrics := map[string]prometheus.Summary{}
	for backendType := range caManagers {
		latencyMetrics[backendType] = prometheus.NewSummary(prometheus.SummaryOpts{
			Name:       "ca_manager_get_root_cert_" + backendType,
			Help:       "Summary of CA manager get CA root certificate latencies",
			Objectives: core_metrics.DefaultObjectives,
		})
		if err := metrics.Register(latencyMetrics[backendType]); err != nil {
			return nil, err
		}
	}
	return &meshCaProvider{
		caManagers:     caManagers,
		latencyMetrics: latencyMetrics,
	}, nil
}

type meshCaProvider struct {
	caManagers core_ca.Managers
	// latencyMetrics maps backend type to backend cert retrieval summary metrics
	latencyMetrics map[string]prometheus.Summary
}

func (s *meshCaProvider) Get(ctx context.Context, mesh *core_mesh.MeshResource) (*core_xds.CaSecret, []string, error) {
	backend := mesh.GetEnabledCertificateAuthorityBackend()
	if backend == nil {
		return nil, nil, errors.New("CA backend is nil")
	}

	var cancel context.CancelFunc
	timeout := backend.GetRootCert().GetRequestTimeout()
	if timeout != nil {
		ctx, cancel = context.WithTimeout(ctx, timeout.AsDuration())
		defer cancel()
	}

	caManager, exist := s.caManagers[backend.Type]
	if !exist {
		return nil, nil, errors.Errorf("CA manager of type %s not exist", backend.Type)
	}

	var certs [][]byte
	var err error
	func() {
		start := time.Now()
		defer func() {
			s.latencyMetrics[backend.Type].Observe(float64(time.Since(start).Milliseconds()))
		}()
		certs, err = caManager.GetRootCert(ctx, mesh.GetMeta().GetName(), backend)
	}()
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not get root certs")
	}

	return &core_xds.CaSecret{
		PemCerts: certs,
	}, []string{backend.Name}, nil
}
