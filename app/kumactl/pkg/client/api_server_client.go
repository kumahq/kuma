package client

import (
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

const (
	// Time limit for requests to the Control Plane API Server.
	Timeout = 60 * time.Second
)

func ApiServerClient(coordinates *config_proto.ControlPlaneCoordinates_ApiServer) (util_http.Client, error) {
	baseURL, err := url.Parse(coordinates.Url)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse API Server URL")
	}
	client := &http.Client{
		Timeout: Timeout,
	}
	if err := util_http.ConfigureMTLS(client, coordinates.CaCertFile, coordinates.ClientCertFile, coordinates.ClientKeyFile); err != nil {
		return nil, errors.Wrap(err, "could not configure HTTP client with TLS")
	}
	return util_http.ClientWithBaseURL(client, baseURL), nil
}
