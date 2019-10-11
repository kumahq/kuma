package tokens

import (
	"bytes"
	"encoding/json"
	"github.com/Kong/kuma/pkg/tokens/builtin/server/model"
	util_http "github.com/Kong/kuma/pkg/util/http"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	timeout = 10 * time.Second
)

func NewDpTokenClient(address string) (DpTokenClient, error) {
	baseURL, err := url.Parse(address)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse API Server URL")
	}
	client := util_http.ClientWithTimeout(util_http.ClientWithBaseURL(&http.Client{}, baseURL), timeout)
	return &httpDpTokenClient{
		client: client,
	}, nil
}

type DpTokenClient interface {
	Generate(name string, mesh string) (string, error)
}

type httpDpTokenClient struct {
	client util_http.Client
}

var _ DpTokenClient = &httpDpTokenClient{}

func (h *httpDpTokenClient) Generate(name string, mesh string) (string, error) {
	tokenReq := &model.DataplaneTokenRequest{
		Name: name,
		Mesh: mesh,
	}
	reqBytes, err := json.Marshal(tokenReq)
	if err != nil {
		return "", errors.Wrap(err, "could not marshal token request to json")
	}
	req, err := http.NewRequest("GET", "/token", bytes.NewReader(reqBytes))
	if err != nil {
		return "", errors.Wrap(err, "could not construct the request")
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "could not execute the request")
	}
	if resp.StatusCode != 200 {
		return "", errors.Errorf("unexpected status code %d. Expected 200", resp.StatusCode)
	}
	tokenBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "could not read a body of the request")
	}
	return string(tokenBytes), nil
}
