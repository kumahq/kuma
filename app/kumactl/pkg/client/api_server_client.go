package client

import (
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

func ApiServerClient(coordinates *config_proto.ControlPlaneCoordinates_ApiServer, timeout time.Duration) (util_http.Client, error) {
	headers := make(map[string]string)
	baseURL, err := url.Parse(coordinates.Url)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse API Server URL")
	}
	client := &http.Client{
		Timeout: timeout,
	}
	if err := util_http.ConfigureMTLS(client, coordinates.CaCertFile, coordinates.ClientCertFile, coordinates.ClientKeyFile); err != nil {
		return nil, errors.Wrap(err, "could not configure HTTP client with TLS")
	}
	for _, h := range coordinates.Headers {
		headers[h.Key] = h.Value
	}
	return util_http.ClientWithBaseURL(client, baseURL, headers), nil
}
