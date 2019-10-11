package resources

import (
	util_http "github.com/Kong/kuma/pkg/util/http"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"time"
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
	return util_http.ClientWithBaseURL(util_http.ClientWithTimeout(&http.Client{}, Timeout), baseURL), nil
}
