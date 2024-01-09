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
	utilnet "github.com/kumahq/kuma/pkg/util/net"
)

type ConfigFetcher struct {
	socketPath string
	ticker     *time.Ticker
	hijacker   *metrics.Hijacker
}

var _ component.Component = &ConfigFetcher{}

var logger = core.Log.WithName("mesh-metric-config-fetcher")

func NewMeshMetricConfigFetcher(socketPath string, ticker *time.Ticker, hijacker *metrics.Hijacker) component.Component {
	return &ConfigFetcher{
		socketPath: socketPath,
		ticker:     ticker,
		hijacker:   hijacker,
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
			configuration, err := httpc.Get("http://localhost/meshmeetric")
			if err != nil {
				logger.Info("failed to scrape /meshmetric endpoint", "err", err)
				continue
			}
			defer configuration.Body.Close()
			conf := Configuration{}

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
			cf.hijacker.SetApplicationsToScrape(mapApplicationToApplicationToScrape(conf.Observability.Metrics.Applications))
		case <-stop:
			logger.Info("stopping Dynamic Mesh Metrics Configuration Scraper")
			return nil
		}
	}
}

func (cf ConfigFetcher) NeedLeaderElection() bool {
	return false
}

type Application struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Address string `json:"address"`
	Port    uint32 `json:"port"`
	Regex   string `json:"regex,omitempty"`
}

type Configuration struct {
	Observability struct {
		Metrics struct {
			Applications []Application `json:"applications"`
			Backends     []struct {
				Type          string `json:"type"`
				OpenTelemetry struct {
					Endpoint string `json:"endpoint"`
				} `json:"openTelemetry,omitempty"`
			} `json:"backends,omitempty"`
		} `json:"metrics"`
	} `json:"observability"`
}

func mapApplicationToApplicationToScrape(applications []Application) []metrics.ApplicationToScrape {
	applicationsToScrape := make([]metrics.ApplicationToScrape, 0, len(applications))

	for i, application := range applications {
		applicationsToScrape[i] = metrics.ApplicationToScrape{
			Name:          application.Name,
			Address:       application.Address,
			Path:          application.Path,
			Port:          application.Port,
			IsIPv6:        utilnet.IsAddressIPv6(application.Address),
			QueryModifier: metrics.RemoveQueryParameters,
		}
	}

	return applicationsToScrape
}
