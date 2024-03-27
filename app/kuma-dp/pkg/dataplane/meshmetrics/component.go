package meshmetrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/metrics"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/plugin/xds"
	utilnet "github.com/kumahq/kuma/pkg/util/net"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type ConfigFetcher struct {
	httpClient            http.Client
	socketPath            string
	ticker                *time.Ticker
	hijacker              *metrics.Hijacker
	defaultAddress        string
	envoyAdminAddress     string
	envoyAdminPort        uint32
	openTelemetryProducer *metrics.AggregatedProducer
	runningBackends       map[string]*runningBackend
	drainTime             time.Duration
}

const unixDomainSocket = "unix"

var _ component.Component = &ConfigFetcher{}

var logger = core.Log.WithName("mesh-metric-config-fetcher")

func NewMeshMetricConfigFetcher(socketPath string, ticker *time.Ticker, hijacker *metrics.Hijacker, openTelemetryProducer *metrics.AggregatedProducer, address string, envoyAdminPort uint32, envoyAdminAddress string, drainTime time.Duration) component.Component {
	return &ConfigFetcher{
		httpClient: http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial(unixDomainSocket, socketPath)
				},
			},
		},
		socketPath:            socketPath,
		ticker:                ticker,
		hijacker:              hijacker,
		openTelemetryProducer: openTelemetryProducer,
		defaultAddress:        address,
		envoyAdminAddress:     envoyAdminAddress,
		envoyAdminPort:        envoyAdminPort,
		runningBackends:       map[string]*runningBackend{},
		drainTime:             drainTime,
	}
}

func (cf *ConfigFetcher) Start(stop <-chan struct{}) error {
	logger.Info("starting Dynamic Mesh Metrics Configuration Scraper",
		"socketPath", fmt.Sprintf("unix://%s", cf.socketPath),
	)

	ctx, ctxCancel := context.WithCancel(context.Background())
	go func() {
		<-stop
		ctxCancel()
	}()

	for {
		select {
		case <-cf.ticker.C:
			if _, err := os.Stat(cf.socketPath); errors.Is(err, os.ErrNotExist) {
				logger.V(1).Info("skipping /meshmetric endpoint scrape since socket does not exist", "err", err)
				continue
			}

			configuration, err := cf.scrapeConfig()
			if err != nil {
				continue
			}
			logger.V(1).Info("updating hijacker configuration", "conf", configuration)
			newApplicationsToScrape := cf.mapApplicationToApplicationToScrape(configuration.Observability.Metrics.Applications, configuration.Observability.Metrics.Sidecar)
			cf.configurePrometheus(newApplicationsToScrape, getPrometheusBackends(configuration.Observability.Metrics.Backends))
			err = cf.configureOpenTelemetryExporter(ctx, newApplicationsToScrape, getOpenTelemetryBackends(configuration.Observability.Metrics.Backends))
			if err != nil {
				logger.Error(err, "Configuring OpenTelemetry Exporter failed")
				continue
			}
		case <-stop:
			logger.Info("stopping Dynamic Mesh Metrics Configuration Scraper")
			cf.shutDownMetricsExporters()
			return nil
		}
	}
}

func (cf *ConfigFetcher) NeedLeaderElection() bool {
	return false
}

func (cf *ConfigFetcher) scrapeConfig() (*xds.MeshMetricDpConfig, error) {
	conf := xds.MeshMetricDpConfig{}
	// since we use socket for communication "localhost" is ignored but this is needed for this
	// http call to work
	configuration, err := cf.httpClient.Get("http://localhost/meshmetric")
	if err != nil {
		// this error can only occur when we configured policy once and then remove it. Listener is removed but socket file
		// is still present since Envoy does not clean it.
		if strings.Contains(err.Error(), "connection refused") {
			logger.V(1).Info("Failed to scrape config, Envoy not listening on socket")
			return &conf, nil
		}
		logger.Info("failed to scrape /meshmetric endpoint", "err", err)
		return nil, errors.Wrap(err, "failed to scrape /meshmetric endpoint")
	}

	defer configuration.Body.Close()
	respBytes, err := io.ReadAll(configuration.Body)
	if err != nil {
		logger.Info("failed to read bytes of the response", "err", err)
		return nil, errors.Wrap(err, "failed to read bytes of the response")
	}
	if err = json.Unmarshal(respBytes, &conf); err != nil {
		logger.Info("failed to unmarshall the response", "err", err)
		return nil, errors.Wrap(err, "failed to unmarshall the response")
	}

	return &conf, nil
}

func (cf *ConfigFetcher) configurePrometheus(applicationsToScrape []metrics.ApplicationToScrape, prometheusBackends []xds.Backend) {
	if len(prometheusBackends) == 0 {
		return
	}
	cf.openTelemetryProducer.SetApplicationsToScrape(applicationsToScrape)
}

func (cf *ConfigFetcher) configureOpenTelemetryExporter(ctx context.Context, applicationsToScrape []metrics.ApplicationToScrape, openTelemetryBackends map[string]*xds.OpenTelemetryBackend) error {
	err := cf.reconfigureBackends(ctx, openTelemetryBackends)
	if err != nil {
		return err
	}
	err = cf.shutdownBackendsRemovedFromConfig(ctx, openTelemetryBackends)
	if err != nil {
		return err
	}
	cf.openTelemetryProducer.SetApplicationsToScrape(applicationsToScrape)
	return nil
}

func (cf *ConfigFetcher) reconfigureBackends(ctx context.Context, openTelemetryBackends map[string]*xds.OpenTelemetryBackend) error {
	for backendName, backend := range openTelemetryBackends {
		// backend already running, in the future we can reconfigure it here
		if cf.runningBackends[backendName] != nil {
			err := cf.reconfigureBackendIfNeeded(ctx, backendName, backend)
			if err != nil {
				return err
			}
			continue
		}
		// start backend as it is not running yet
		exporter, err := startExporter(ctx, backend, cf.openTelemetryProducer, backendName)
		if err != nil {
			return err
		}
		cf.runningBackends[backendName] = exporter
	}
	return nil
}

func (cf *ConfigFetcher) shutdownBackendsRemovedFromConfig(ctx context.Context, openTelemetryBackends map[string]*xds.OpenTelemetryBackend) error {
	var backendsToRemove []string
	for backendName := range cf.runningBackends {
		// backend still configured in policy
		if openTelemetryBackends[backendName] != nil {
			continue
		}
		backendsToRemove = append(backendsToRemove, backendName)
	}
	for _, backendName := range backendsToRemove {
		logger.Info("Shutting down OpenTelemetry exporter", "backend", backendName)
		err := cf.shutdownBackend(ctx, backendName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cf *ConfigFetcher) shutdownBackend(ctx context.Context, backendName string) error {
	err := cf.runningBackends[backendName].exporter.Shutdown(ctx)
	if err != nil && !errors.Is(err, sdkmetric.ErrReaderShutdown) {
		return err
	}
	delete(cf.runningBackends, backendName)
	return nil
}

func startExporter(ctx context.Context, backend *xds.OpenTelemetryBackend, producer *metrics.AggregatedProducer, backendName string) (*runningBackend, error) {
	if backend == nil {
		return nil, nil
	}
	logger.Info("Starting OpenTelemetry exporter", "backend", backendName)
	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(backend.Endpoint),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithDialOption(dialOptions()...),
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

func getOpenTelemetryBackends(allBackends []xds.Backend) map[string]*xds.OpenTelemetryBackend {
	openTelemetryBackends := map[string]*xds.OpenTelemetryBackend{}
	for _, backend := range allBackends {
		if backend.Type == string(v1alpha1.OpenTelemetryBackendType) {
			openTelemetryBackends[pointer.Deref(backend.Name)] = backend.OpenTelemetry
		}
	}
	return openTelemetryBackends
}

func getPrometheusBackends(allBackends []xds.Backend) []xds.Backend {
	var prometheusBackends []xds.Backend
	for _, backend := range allBackends {
		if backend.Type == string(v1alpha1.PrometheusBackendType) {
			prometheusBackends = append(prometheusBackends, backend)
		}
	}
	return prometheusBackends
}

func (cf *ConfigFetcher) mapApplicationToApplicationToScrape(applications []xds.Application, sidecar *v1alpha1.Sidecar) []metrics.ApplicationToScrape {
	var applicationsToScrape []metrics.ApplicationToScrape

	for _, application := range applications {
		address := cf.defaultAddress
		if application.Address != "" {
			address = application.Address
		}
		applicationsToScrape = append(applicationsToScrape, metrics.ApplicationToScrape{
			Name:              pointer.Deref(application.Name),
			Address:           address,
			Path:              application.Path,
			Port:              application.Port,
			IsIPv6:            utilnet.IsAddressIPv6(address),
			QueryModifier:     metrics.RemoveQueryParameters,
			MeshMetricMutator: metrics.AggregatedOtelMutator(),
		})
	}

	applicationsToScrape = append(applicationsToScrape, metrics.ApplicationToScrape{
		Name:              "envoy",
		Path:              "/stats",
		Address:           cf.envoyAdminAddress,
		Port:              cf.envoyAdminPort,
		IsIPv6:            false,
		QueryModifier:     metrics.AggregatedQueryParametersModifier(metrics.AddPrometheusFormat, metrics.AddSidecarParameters(sidecar)),
		Mutator:           metrics.AggregatedMetricsMutator(metrics.MergeClustersForPrometheus),
		MeshMetricMutator: metrics.AggregatedOtelMutator(metrics.ProfileMutatorGenerator(sidecar), metrics.MergeClustersForOpenTelemetry),
	})

	return applicationsToScrape
}

func (cf *ConfigFetcher) reconfigureBackendIfNeeded(ctx context.Context, backendName string, backend *xds.OpenTelemetryBackend) error {
	if configChanged(cf.runningBackends[backendName].appliedConfig, backend) {
		err := cf.shutdownBackend(ctx, backendName)
		if err != nil {
			return err
		}
		exporter, err := startExporter(ctx, backend, cf.openTelemetryProducer, backendName)
		if err != nil {
			return err
		}
		cf.runningBackends[backendName] = exporter
	}
	return nil
}

func (cf *ConfigFetcher) shutDownMetricsExporters() {
	ctx, cancel := context.WithTimeout(context.Background(), cf.drainTime)
	defer cancel()
	for backendName := range cf.runningBackends {
		err := cf.shutdownBackend(ctx, backendName)
		if err != nil {
			logger.Error(err, "Failed shutting down metric exporter")
		}
	}
}

func configChanged(appliedConfig xds.OpenTelemetryBackend, newConfig *xds.OpenTelemetryBackend) bool {
	return appliedConfig.RefreshInterval.Duration != newConfig.RefreshInterval.Duration
}

func dialOptions() []grpc.DialOption {
	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, unixDomainSocket, addr)
	}
	return []grpc.DialOption{
		grpc.WithContextDialer(dialer),
	}
}

type runningBackend struct {
	exporter      *sdkmetric.MeterProvider
	appliedConfig xds.OpenTelemetryBackend
}
