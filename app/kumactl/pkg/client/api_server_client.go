package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

func ApiServerClient(coordinates *config_proto.ControlPlaneCoordinates_ApiServer, timeout time.Duration) (util_http.Client, error) {
	headers := make(map[string]string)
	baseURL, err := url.Parse(coordinates.Url)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse API Server URL: %w", err)
	}
	client := &http.Client{
		Timeout: timeout,
	}
	if err := util_http.ConfigureMTLS(client, coordinates.CaCertFile, coordinates.ClientCertFile, coordinates.ClientKeyFile); err != nil {
		return nil, fmt.Errorf("could not configure HTTP client with TLS: %w", err)
	}
	for _, h := range coordinates.Headers {
		headers[h.Key] = h.Value
	}
	return util_http.ClientWithBaseURL(client, baseURL, headers), nil
}
