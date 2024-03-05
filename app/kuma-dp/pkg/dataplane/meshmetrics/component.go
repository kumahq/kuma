package meshmetrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
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
}

const unixDomainSocket = "unix"

var _ component.Component = &ConfigFetcher{}

var logger = core.Log.WithName("mesh-metric-config-fetcher")

func NewMeshMetricConfigFetcher(socketPath string, ticker *time.Ticker, hijacker *metrics.Hijacker, openTelemetryProducer *metrics.AggregatedProducer, address string, envoyAdminPort uint32, envoyAdminAddress string) component.Component {
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
	}
}

func (cf *ConfigFetcher) Start(stop <-chan struct{}) error {
	logger.Info("starting Dynamic Mesh Metrics Configuration Scraper",
		"socketPath", fmt.Sprintf("unix://%s", cf.socketPath),
	)

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
			err = cf.configureOpenTelemetryExporter(newApplicationsToScrape, getOpenTelemetryBackends(configuration.Observability.Metrics.Backends))
			if err != nil {
				logger.Error(err, "Configuring OpenTelemetry Exporter failed")
				continue
			}
		case <-stop:
			logger.Info("stopping Dynamic Mesh Metrics Configuration Scraper")
			return nil
		}
	}
}

func (cf *ConfigFetcher) NeedLeaderElection() bool {
	return false
}

func (cf *ConfigFetcher) scrapeConfig() (*xds.MeshMetricDpConfig, error) {
	// since we use socket for communication "localhost" is ignored but this is needed for this
	// http call to work
	configuration, err := cf.httpClient.Get("http://localhost/meshmetric")
	if err != nil {
		logger.Info("failed to scrape /meshmetric endpoint", "err", err)
		return nil, errors.Wrap(err, "failed to scrape /meshmetric endpoint")
	}

	defer configuration.Body.Close()
	conf := xds.MeshMetricDpConfig{}

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

func (cf *ConfigFetcher) configureOpenTelemetryExporter(applicationsToScrape []metrics.ApplicationToScrape, openTelemetryBackends map[string]*xds.OpenTelemetryBackend) error {
	err := cf.reconfigureBackends(openTelemetryBackends)
	if err != nil {
		return err
	}
	err = cf.shutdownBackendsRemovedFromConfig(openTelemetryBackends)
	if err != nil {
		return err
	}
	cf.openTelemetryProducer.SetApplicationsToScrape(applicationsToScrape)
	return nil
}

func (cf *ConfigFetcher) reconfigureBackends(openTelemetryBackends map[string]*xds.OpenTelemetryBackend) error {
	for backendName, backend := range openTelemetryBackends {
		// backend already running, in the future we can reconfigure it here
		if cf.runningBackends[backendName] != nil {
			err := cf.reconfigureBackendIfNeeded(backendName, backend)
			if err != nil {
				return err
			}
		}
		// start backend as it is not running yet
		exporter, err := startExporter(backend, cf.openTelemetryProducer, backendName)
		if err != nil {
			return err
		}
		cf.runningBackends[backendName] = exporter
	}
	return nil
}

func (cf *ConfigFetcher) shutdownBackendsRemovedFromConfig(openTelemetryBackends map[string]*xds.OpenTelemetryBackend) error {
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
		err := cf.shutdownBackend(backendName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cf *ConfigFetcher) shutdownBackend(backendName string) error {
	err := cf.runningBackends[backendName].exporter.Shutdown(context.Background())
	if err != nil {
		return err
	}
	delete(cf.runningBackends, backendName)
	return nil
}

func startExporter(backend *xds.OpenTelemetryBackend, producer *metrics.AggregatedProducer, backendName string) (*runningBackend, error) {
	if backend == nil {
		return nil, nil
	}
	logger.Info("Starting OpenTelemetry exporter", "backend", backendName)
	exporter, err := otlpmetricgrpc.New(
		context.Background(),
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
			Name:          pointer.Deref(application.Name),
			Address:       address,
			Path:          application.Path,
			Port:          application.Port,
			IsIPv6:        utilnet.IsAddressIPv6(address),
			QueryModifier: metrics.RemoveQueryParameters,
			OtelMutator:   metrics.ParsePrometheusMetrics,
		})
	}

	applicationsToScrape = append(applicationsToScrape, metrics.ApplicationToScrape{
		Name:          "envoy",
		Path:          "/stats",
		Address:       cf.envoyAdminAddress,
		Port:          cf.envoyAdminPort,
		IsIPv6:        false,
		QueryModifier: metrics.AggregatedQueryParametersModifier(metrics.AddPrometheusFormat, metrics.AddSidecarParameters(sidecar)),
		Mutator:       metrics.MergeClusters,
		OtelMutator:   metrics.MergeClustersForOpenTelemetry,
	})

	return applicationsToScrape
}

func (cf *ConfigFetcher) reconfigureBackendIfNeeded(backendName string, backend *xds.OpenTelemetryBackend) error {
	if configChanged(cf.runningBackends[backendName].appliedConfig, backend) {
		err := cf.shutdownBackend(backendName)
		if err != nil {
			return err
		}
		exporter, err := startExporter(backend, cf.openTelemetryProducer, backendName)
		if err != nil {
			return err
		}
		cf.runningBackends[backendName] = exporter
	}
	return nil
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
