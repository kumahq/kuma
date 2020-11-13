package tokens

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	kumactl_config "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	error_types "github.com/kumahq/kuma/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/pkg/tokens/builtin/server/types"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

const (
	timeout = 10 * time.Second
)

func NewDataplaneTokenClient(config *kumactl_config.ControlPlaneCoordinates_ApiServer) (DataplaneTokenClient, error) {
	baseURL, err := url.Parse(config.Url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse Dataplane Token Server URL")
	}
	httpClient := &http.Client{
		Timeout: timeout,
	}
	if baseURL.Scheme == "https" {
		if err := util_http.ConfigureMTLS(httpClient, config.CaCertFile, config.ClientCertFile, config.ClientKeyFile); err != nil {
			return nil, errors.Wrap(err, "could not configure tls for dataplane token client")
		}
	}
	client := util_http.ClientWithBaseURL(httpClient, baseURL)
	return &httpDataplaneTokenClient{
		client: client,
	}, nil
}

type DataplaneTokenClient interface {
	Generate(name string, mesh string, tags map[string][]string, dpType string) (string, error)
}

type httpDataplaneTokenClient struct {
	client util_http.Client
}

var _ DataplaneTokenClient = &httpDataplaneTokenClient{}

func (h *httpDataplaneTokenClient) Generate(name string, mesh string, tags map[string][]string, dpType string) (string, error) {
	tokenReq := &types.DataplaneTokenRequest{
		Name: name,
		Mesh: mesh,
		Tags: tags,
		Type: dpType,
	}
	reqBytes, err := json.Marshal(tokenReq)
	if err != nil {
		return "", errors.Wrap(err, "could not marshal token request to json")
	}
	req, err := http.NewRequest("POST", "/tokens", bytes.NewReader(reqBytes))
	if err != nil {
		return "", errors.Wrap(err, "could not construct the request")
	}
	req.Header.Set("content-type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "could not execute the request")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "could not read a body of the request")
	}
	if resp.StatusCode != 200 {
		kumaErr := error_types.Error{}
		if err := json.Unmarshal(body, &kumaErr); err == nil {
			if kumaErr.Title != "" && kumaErr.Details != "" {
				return "", &kumaErr
			}
		}
		return "", errors.Errorf("(%d): %s", resp.StatusCode, body)
	}
	return string(body), nil
}
