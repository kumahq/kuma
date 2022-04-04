package secrets

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

type Identity struct {
	Mesh     string
	Name     string
	Services mesh_proto.MultiValueTagSet
}

type IdentityProvider interface {
	// Get returns PEM encoded cert + key, backend that was used to generate this pair and an error.
	Get(context.Context, Identity, *core_mesh.MeshResource) (*core_xds.IdentitySecret, string, error)
}

func NewIdentityProvider(caManagers core_ca.Managers, metrics core_metrics.Metrics) (IdentityProvider, error) {
	latencyMetrics := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "ca_manager_get_cert",
		Help:       "Summary of CA manager get certificate latencies",
		Objectives: core_metrics.DefaultObjectives,
	}, []string{"backend_name"})
	if err := metrics.Register(latencyMetrics); err != nil {
		return nil, err
	}
	return &identityCertProvider{
		caManagers:     caManagers,
		latencyMetrics: latencyMetrics,
	}, nil
}

type identityCertProvider struct {
	caManagers     core_ca.Managers
	latencyMetrics *prometheus.SummaryVec
}

func (s *identityCertProvider) Get(ctx context.Context, requestor Identity, mesh *core_mesh.MeshResource) (*core_xds.IdentitySecret, string, error) {
	backend := mesh.GetEnabledCertificateAuthorityBackend()
	if backend == nil {
		return nil, "", errors.Errorf("CA default backend in mesh %q has to be defined", mesh.GetMeta().GetName())
	}

	timeout := backend.GetDpCert().GetRequestTimeout()
	if timeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout.AsDuration())
		defer cancel()
	}

	caManager, exist := s.caManagers[backend.Type]
	if !exist {
		return nil, "", errors.Errorf("CA manager of type %s not exist", backend.Type)
	}

	var pair core_ca.KeyPair
	var err error
	func() {
		start := time.Now()
		defer func() {
			s.latencyMetrics.WithLabelValues(backend.GetName()).Observe(float64(time.Since(start).Milliseconds()))
		}()
		pair, err = caManager.GenerateDataplaneCert(ctx, mesh.GetMeta().GetName(), backend, requestor.Services)
	}()

	if err != nil {
		return nil, "", errors.Wrapf(err, "could not generate dataplane cert for mesh: %q backend: %q services: %q", mesh.GetMeta().GetName(), backend.Name, requestor.Services)
	}

	return &core_xds.IdentitySecret{
		PemCerts: [][]byte{pair.CertPEM},
		PemKey:   pair.KeyPEM,
	}, backend.Name, nil
}
