package meshmetrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/metrics"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/dpapi"
	utilnet "github.com/kumahq/kuma/pkg/util/net"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type Manager struct {
	hijacker              *metrics.Hijacker
	defaultAddress        string
	envoyAdminAddress     string
	envoyAdminPort        uint32
	openTelemetryProducer *metrics.AggregatedProducer
	runningBackends       map[string]*runningBackend
	drainTime             time.Duration
	ctx                   context.Context
	cancel                context.CancelFunc
	done                  chan struct{}
	newConfig             chan dpapi.MeshMetricDpConfig
}

var logger = core.Log.WithName("mesh-metric-config-fetcher")

var _ component.GracefulComponent = &Manager{}

func NewManager(ctx context.Context, hijacker *metrics.Hijacker, openTelemetryProducer *metrics.AggregatedProducer, address string, envoyAdminPort uint32, envoyAdminAddress string, drainTime time.Duration) *Manager {
	ctx, cancel := context.WithCancel(ctx)
	return &Manager{
		ctx:                   ctx,
		cancel:                cancel,
		hijacker:              hijacker,
		openTelemetryProducer: openTelemetryProducer,
		defaultAddress:        address,
		envoyAdminAddress:     envoyAdminAddress,
		envoyAdminPort:        envoyAdminPort,
		runningBackends:       map[string]*runningBackend{},
		drainTime:             drainTime,
		newConfig:             make(chan dpapi.MeshMetricDpConfig),
		done:                  make(chan struct{}),
	}
}

func (m *Manager) Start(stop <-chan struct{}) error {
	defer close(m.done)
	for {
		select {
		case configuration := <-m.newConfig:
			logger.Info("updating hijacker configuration", "conf", configuration)
			err := m.Step(configuration)
			if err != nil {
				logger.Error(err, "failed to update hijacker configuration")
			}
		case <-stop:
			return m.Shutdown()
		}
	}
}

func (m *Manager) NeedLeaderElection() bool {
	return false
}

func (m *Manager) WaitForDone() {
	<-m.done
}

func (m *Manager) OnChange(ctx context.Context, reader io.Reader) error {
	configuration := dpapi.MeshMetricDpConfig{}
	if err := json.NewDecoder(reader).Decode(&configuration); err != nil {
		return fmt.Errorf("mesh metric configuration decoding error: %w", err)
	}
	select {
	case m.newConfig <- configuration:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (m *Manager) configurePrometheus(applicationsToScrape []metrics.ApplicationToScrape, prometheusBackends []dpapi.Backend) {
	if len(prometheusBackends) == 0 {
		return
	}
	m.openTelemetryProducer.SetApplicationsToScrape(applicationsToScrape)
}

func (m *Manager) configureOpenTelemetryExporter(ctx context.Context, applicationsToScrape []metrics.ApplicationToScrape, openTelemetryBackends map[string]*dpapi.OpenTelemetryBackend) error {
	err := m.reconfigureBackends(ctx, openTelemetryBackends)
	if err != nil {
		return err
	}
	err = m.shutdownBackendsRemovedFromConfig(ctx, openTelemetryBackends)
	if err != nil {
		return err
	}
	m.openTelemetryProducer.SetApplicationsToScrape(applicationsToScrape)
	return nil
}

func (m *Manager) reconfigureBackends(ctx context.Context, openTelemetryBackends map[string]*dpapi.OpenTelemetryBackend) error {
	for backendName, backend := range openTelemetryBackends {
		// backend already running, in the future we can reconfigure it here
		if m.runningBackends[backendName] != nil {
			err := m.reconfigureBackend(ctx, backendName, backend)
			if err != nil {
				return err
			}
			continue
		}
		// start backend as it is not running yet
		exporter, err := startExporter(ctx, backend, m.openTelemetryProducer, backendName)
		if err != nil {
			return err
		}
		m.runningBackends[backendName] = exporter
	}
	return nil
}

func (m *Manager) shutdownBackendsRemovedFromConfig(ctx context.Context, openTelemetryBackends map[string]*dpapi.OpenTelemetryBackend) error {
	var backendsToRemove []string
	for backendName := range m.runningBackends {
		// backend still configured in policy
		if openTelemetryBackends[backendName] != nil {
			continue
		}
		backendsToRemove = append(backendsToRemove, backendName)
	}
	for _, backendName := range backendsToRemove {
		logger.Info("Shutting down OpenTelemetry exporter", "backend", backendName)
		err := m.shutdownBackend(ctx, backendName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) shutdownBackend(ctx context.Context, backendName string) error {
	err := m.runningBackends[backendName].exporter.Shutdown(ctx)
	if err != nil && !errors.Is(err, sdkmetric.ErrReaderShutdown) {
		return err
	}
	delete(m.runningBackends, backendName)
	return nil
}

func startExporter(ctx context.Context, backend *dpapi.OpenTelemetryBackend, producer *metrics.AggregatedProducer, backendName string) (*runningBackend, error) {
	if backend == nil {
		return nil, nil
	}
	logger.Info("Starting OpenTelemetry exporter", "backend", backendName)
	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(fmt.Sprintf("unix://%s", backend.Endpoint)),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	return &runningBackend{
		exporter: sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(backend.RefreshInterval.Duration),
				sdkmetric.WithProducer(producer),
			)),
		),
		appliedConfig: *backend,
	}, nil
}

func getOpenTelemetryBackends(allBackends []dpapi.Backend) map[string]*dpapi.OpenTelemetryBackend {
	openTelemetryBackends := map[string]*dpapi.OpenTelemetryBackend{}
	for _, backend := range allBackends {
		if backend.Type == string(v1alpha1.OpenTelemetryBackendType) {
			openTelemetryBackends[pointer.Deref(backend.Name)] = backend.OpenTelemetry
		}
	}
	return openTelemetryBackends
}

func getPrometheusBackends(allBackends []dpapi.Backend) []dpapi.Backend {
	var prometheusBackends []dpapi.Backend
	for _, backend := range allBackends {
		if backend.Type == string(v1alpha1.PrometheusBackendType) {
			prometheusBackends = append(prometheusBackends, backend)
		}
	}
	return prometheusBackends
}

func (m *Manager) mapApplicationToApplicationToScrape(applications []dpapi.Application, sidecar *v1alpha1.Sidecar, extraLabels map[string]string) []metrics.ApplicationToScrape {
	var applicationsToScrape []metrics.ApplicationToScrape

	for _, application := range applications {
		address := m.defaultAddress
		if application.Address != "" {
			address = application.Address
		}
		applicationsToScrape = append(applicationsToScrape, metrics.ApplicationToScrape{
			Name:              pointer.Deref(application.Name),
			Address:           address,
			Path:              application.Path,
			Port:              application.Port,
			IsIPv6:            utilnet.IsAddressIPv6(address),
			ExtraLabels:       extraLabels,
			QueryModifier:     metrics.RemoveQueryParameters,
			MeshMetricMutator: metrics.AggregatedOtelMutator(),
		})
	}

	applicationsToScrape = append(applicationsToScrape, metrics.ApplicationToScrape{
		Name:              "envoy",
		Path:              "/stats",
		Address:           m.envoyAdminAddress,
		Port:              m.envoyAdminPort,
		IsIPv6:            false,
		ExtraLabels:       extraLabels,
		QueryModifier:     metrics.AggregatedQueryParametersModifier(metrics.AddPrometheusFormat, metrics.AddSidecarParameters(sidecar)),
		MeshMetricMutator: metrics.AggregatedOtelMutator(metrics.ProfileMutatorGenerator(sidecar)),
	})

	return applicationsToScrape
}

func (m *Manager) reconfigureBackend(ctx context.Context, backendName string, backend *dpapi.OpenTelemetryBackend) error {
	err := m.shutdownBackend(ctx, backendName)
	if err != nil {
		return err
	}
	exporter, err := startExporter(ctx, backend, m.openTelemetryProducer, backendName)
	if err != nil {
		return err
	}
	m.runningBackends[backendName] = exporter
	return nil
}

func (m *Manager) Step(configuration dpapi.MeshMetricDpConfig) error {
	newApplicationsToScrape := m.mapApplicationToApplicationToScrape(configuration.Observability.Metrics.Applications, configuration.Observability.Metrics.Sidecar, configuration.Observability.Metrics.ExtraLabels)
	m.configurePrometheus(newApplicationsToScrape, getPrometheusBackends(configuration.Observability.Metrics.Backends))
	return m.configureOpenTelemetryExporter(m.ctx, newApplicationsToScrape, getOpenTelemetryBackends(configuration.Observability.Metrics.Backends)) // nolint:contextcheck
}

func (m *Manager) Shutdown() error {
	m.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), m.drainTime)
	defer cancel()
	hasError := false
	for backendName := range m.runningBackends {
		bErr := m.shutdownBackend(ctx, backendName)
		if bErr != nil {
			logger.Error(bErr, "Failed to shutdown backend", "backend", backendName)
			hasError = true
		}
	}
	if hasError {
		return errors.New("failed to shutdown some backend")
	}
	return nil
}

type runningBackend struct {
	exporter      *sdkmetric.MeterProvider
	appliedConfig dpapi.OpenTelemetryBackend
}
