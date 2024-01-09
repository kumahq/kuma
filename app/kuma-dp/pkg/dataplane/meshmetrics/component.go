package meshmetrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

type ConfigFetcher struct {
	socketPath    string
	ticker        *time.Ticker
	Configuration Configuration
}

var _ component.Component = &ConfigFetcher{}

var logger = core.Log.WithName("mesh-metric-config-fetcher")

func NewMeshMetricConfigFetcher(socketPath string, ticker *time.Ticker) component.Component {
	return &ConfigFetcher{
		socketPath: socketPath,
		ticker:     ticker,
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
			configuration, err := httpc.Get("/meshmeetric")
			if err != nil {
				logger.Info("failed to scrape /meshmetric endpoint", "err", err)
			}
			defer configuration.Body.Close()
			conf := Configuration{}

			respBytes, err := io.ReadAll(configuration.Body)
			if err != nil {
				logger.Info("failed to read bytes of the response", "err", err)
			}
			if err = json.Unmarshal(respBytes, &conf); err != nil {
				logger.Info("failed to unmarshall the response", "err", err)
			}

			logger.V(1).Info("updating configuration", "conf", conf)
			cf.Configuration = conf
		case <-stop:
			logger.Info("stopping Dynamic Mesh Metrics Configuration Scraper")
			return nil
		}
	}
}

func (cf ConfigFetcher) NeedLeaderElection() bool {
	return false
}

type Configuration struct {
	Observability struct {
		Metrics struct {
			Applications []struct {
				Path  string `json:"path,omitempty"`
				Port  string `json:"port"`
				Regex string `json:"regex,omitempty"`
			} `json:"applications"`
			Backends []struct {
				Type          string `json:"type"`
				OpenTelemetry struct {
					Endpoint string `json:"endpoint"`
				} `json:"openTelemetry,omitempty"`
			} `json:"backends"`
		} `json:"metrics"`
	} `json:"observability"`
}
