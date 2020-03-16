package resources

import (
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	util_http "github.com/Kong/kuma/pkg/util/http"
)

const (
	// Time limit for requests to the Control Plane API Server.
	Timeout = 60 * time.Second
)

func apiServerClient(apiUrl string) (util_http.Client, error) {
	baseURL, err := url.Parse(apiUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse API Server URL")
	}
	client := &http.Client{
		Timeout: Timeout,
	}
	return util_http.ClientWithBaseURL(client, baseURL), nil
}
