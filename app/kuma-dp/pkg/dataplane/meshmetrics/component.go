package meshmetrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/metrics"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/plugin/xds"
	utilnet "github.com/kumahq/kuma/pkg/util/net"
)

type ConfigFetcher struct {
	socketPath     string
	ticker         *time.Ticker
	hijacker       *metrics.Hijacker
	defaultAddress string
	envoyAdminPort uint32
}

var _ component.Component = &ConfigFetcher{}

var logger = core.Log.WithName("mesh-metric-config-fetcher")

func NewMeshMetricConfigFetcher(socketPath string, ticker *time.Ticker, hijacker *metrics.Hijacker, address string, envoyAdminPort uint32) component.Component {
	return &ConfigFetcher{
		socketPath:     socketPath,
		ticker:         ticker,
		hijacker:       hijacker,
		defaultAddress: address,
		envoyAdminPort: envoyAdminPort,
	}
}

func (cf *ConfigFetcher) Start(stop <-chan struct{}) error {
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", cf.socketPath)
			},
		},
	}

	logger.Info("starting Dynamic Mesh Metrics Configuration Scraper",
		"socketPath", fmt.Sprintf("unix://%s", cf.socketPath),
	)

	for {
		select {
		case <-cf.ticker.C:
			configuration, err := httpc.Get("http://localhost/meshmetric")
			if err != nil {
				logger.Info("failed to scrape /meshmetric endpoint", "err", err)
				continue
			}
			// TODO this defer probably wont work as expected
			defer configuration.Body.Close()
			conf := xds.MeshMetricDpConfig{}

			respBytes, err := io.ReadAll(configuration.Body)
			if err != nil {
				logger.Info("failed to read bytes of the response", "err", err)
				continue
			}
			if err = json.Unmarshal(respBytes, &conf); err != nil {
				logger.Info("failed to unmarshall the response", "err", err)
				continue
			}

			logger.V(1).Info("updating hijacker configuration", "conf", conf)
			cf.hijacker.SetApplicationsToScrape(cf.mapApplicationToApplicationToScrape(conf.Observability.Metrics.Applications))
		case <-stop:
			logger.Info("stopping Dynamic Mesh Metrics Configuration Scraper")
			return nil
		}
	}
}

func (cf *ConfigFetcher) NeedLeaderElection() bool {
	return false
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
		Address:       "127.0.0.1",
		Port:          cf.envoyAdminPort,
		IsIPv6:        false,
		QueryModifier: metrics.AddPrometheusFormat,
		Mutator:       metrics.MergeClusters,
	})

	return applicationsToScrape
}
