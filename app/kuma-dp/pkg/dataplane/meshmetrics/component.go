package meshmetrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/kumahq/kuma/v2/app/kuma-dp/pkg/dataplane/metrics"
	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/dpapi"
	utilnet "github.com/kumahq/kuma/v2/pkg/util/net"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

// OtelExportTarget describes an OTEL metrics export destination received from
// the otelreceiver post-reconcile callback. The otelreceiver.Manager owns the
// Unix socket gRPC server; meshmetrics creates an OTLP metric SDK exporter
// that dials it.
type OtelExportTarget struct {
	Name            string
	SocketPath      string
	RefreshInterval time.Duration
}

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
	newOtelTargets        chan []OtelExportTarget
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
		newOtelTargets:        make(chan []OtelExportTarget, 1),
		done:                  make(chan struct{}),
	}
}

func (m *Manager) Start(stop <-chan struct{}) error {
	defer close(m.done)
	for {
		select {
		case configuration := <-m.newConfig:
			logger.Info("updating hijacker configuration", "conf", configuration)
			m.stepScraping(configuration)
		case targets := <-m.newOtelTargets:
			logger.Info("updating OTEL export targets", "count", len(targets))
			if err := m.stepOtelExport(targets); err != nil {
				logger.Error(err, "failed to update OTEL export targets")
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

// OnOtelTargetsChange is called by the otelreceiver post-reconcile callback
// (wired in run.go) to push OTEL metric export targets into the Manager.
func (m *Manager) OnOtelTargetsChange(targets []OtelExportTarget) {
	select {
	case m.newOtelTargets <- targets:
	default:
		// Drop stale value, push new one.
		select {
		case <-m.newOtelTargets:
		default:
		}
		m.newOtelTargets <- targets
	}
}

func (m *Manager) stepScraping(configuration dpapi.MeshMetricDpConfig) {
	newApplicationsToScrape := m.mapApplicationToApplicationToScrape(
		configuration.Observability.Metrics.Applications,
		configuration.Observability.Metrics.Sidecar,
		configuration.Observability.Metrics.ExtraLabels,
	)
	m.openTelemetryProducer.SetApplicationsToScrape(newApplicationsToScrape)
}

func (m *Manager) stepOtelExport(targets []OtelExportTarget) error {
	desired := map[string]OtelExportTarget{}
	for _, t := range targets {
		desired[t.Name] = t
	}

	// shutdown removed backends
	var toRemove []string
	for name := range m.runningBackends {
		if _, ok := desired[name]; !ok {
			toRemove = append(toRemove, name)
		}
	}
	for _, name := range toRemove {
		logger.Info("Shutting down OpenTelemetry exporter", "backend", name)
		if err := m.shutdownBackend(m.ctx, name); err != nil {
			logger.Error(err, "failed to shut down OpenTelemetry exporter", "backend", name)
		}
	}

	// start or reconfigure backends
	for name, target := range desired {
		if existing, ok := m.runningBackends[name]; ok {
			if existing.appliedConfig == target {
				continue
			}
			if err := m.shutdownBackend(m.ctx, name); err != nil {
				logger.Error(err, "failed to shut down OpenTelemetry exporter for reconfigure", "backend", name)
				continue
			}
		}
		exporter, err := startExporter(m.ctx, target, m.openTelemetryProducer)
		if err != nil {
			return err
		}
		m.runningBackends[name] = exporter
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

func startExporter(ctx context.Context, target OtelExportTarget, producer *metrics.AggregatedProducer) (*runningBackend, error) {
	logger.Info("Starting OpenTelemetry exporter", "backend", target.Name, "socketPath", target.SocketPath)
	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(fmt.Sprintf("unix://%s", target.SocketPath)),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	readerOpts := []sdkmetric.PeriodicReaderOption{
		sdkmetric.WithInterval(target.RefreshInterval),
	}
	if producer != nil {
		readerOpts = append(readerOpts, sdkmetric.WithProducer(producer))
	}
	return &runningBackend{
		exporter: sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
				exporter,
				readerOpts...,
			)),
		),
		appliedConfig: target,
	}, nil
}

func (m *Manager) mapApplicationToApplicationToScrape(applications []dpapi.Application, sidecar *v1alpha1.Sidecar, extraLabels map[string]string) []metrics.ApplicationToScrape {
	var applicationsToScrape []metrics.ApplicationToScrape
	extraAttributes := mapToAttributes(extraLabels)

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
			ExtraAttributes:   extraAttributes,
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
		ExtraAttributes:   extraAttributes,
		QueryModifier:     metrics.AggregatedQueryParametersModifier(metrics.AddPrometheusFormat, metrics.AddSidecarParameters(sidecar)),
		MeshMetricMutator: metrics.AggregatedOtelMutator(metrics.ProfileMutatorGenerator(sidecar)),
	})

	return applicationsToScrape
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

func mapToAttributes(extraLabels map[string]string) []attribute.KeyValue {
	var extraAttributes []attribute.KeyValue
	for k, v := range extraLabels {
		extraAttributes = append(extraAttributes, attribute.String(k, v))
	}
	return extraAttributes
}

type runningBackend struct {
	exporter      *sdkmetric.MeterProvider
	appliedConfig OtelExportTarget
}
