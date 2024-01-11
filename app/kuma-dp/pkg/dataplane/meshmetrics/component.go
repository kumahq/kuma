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

	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/metrics"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/plugin/xds"
	utilnet "github.com/kumahq/kuma/pkg/util/net"
)

type ConfigFetcher struct {
	httpClient        http.Client
	socketPath        string
	ticker            *time.Ticker
	hijacker          *metrics.Hijacker
	defaultAddress    string
	envoyAdminAddress string
	envoyAdminPort    uint32
}

var _ component.Component = &ConfigFetcher{}

var logger = core.Log.WithName("mesh-metric-config-fetcher")

func NewMeshMetricConfigFetcher(socketPath string, ticker *time.Ticker, hijacker *metrics.Hijacker, address string, envoyAdminPort uint32, envoyAdminAddress string) component.Component {
	return &ConfigFetcher{
		httpClient: http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", socketPath)
				},
			},
		},
		socketPath:        socketPath,
		ticker:            ticker,
		hijacker:          hijacker,
		defaultAddress:    address,
		envoyAdminAddress: envoyAdminAddress,
		envoyAdminPort:    envoyAdminPort,
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
			cf.hijacker.SetApplicationsToScrape(cf.mapApplicationToApplicationToScrape(configuration.Observability.Metrics.Applications))
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

func (cf *ConfigFetcher) mapApplicationToApplicationToScrape(applications []xds.Application) []metrics.ApplicationToScrape {
	var applicationsToScrape []metrics.ApplicationToScrape

	for _, application := range applications {
		address := cf.defaultAddress
		if application.Address != "" {
			address = application.Address
		}
		applicationsToScrape = append(applicationsToScrape, metrics.ApplicationToScrape{
			Address:       address,
			Path:          application.Path,
			Port:          application.Port,
			IsIPv6:        utilnet.IsAddressIPv6(application.Address),
			QueryModifier: metrics.RemoveQueryParameters,
		})
	}

	applicationsToScrape = append(applicationsToScrape, metrics.ApplicationToScrape{
		Name:          "envoy",
		Path:          "/stats",
		Address:       cf.envoyAdminAddress,
		Port:          cf.envoyAdminPort,
		IsIPv6:        false,
		QueryModifier: metrics.AddPrometheusFormat,
		Mutator:       metrics.MergeClusters,
	})

	return applicationsToScrape
}
