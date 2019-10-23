package client

import (
	"encoding/json"
	"github.com/Kong/kuma/pkg/coordinates"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	util_http "github.com/Kong/kuma/pkg/util/http"
)

const (
	timeout = 10 * time.Second
)

type CoordinatesClient interface {
	Coordinates() (coordinates.Coordinates, error)
}

func NewCoordinatesClient(address string) (CoordinatesClient, error) {
	baseURL, err := url.Parse(address)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse API Server URL")
	}
	client := util_http.ClientWithTimeout(util_http.ClientWithBaseURL(&http.Client{}, baseURL), timeout)
	return &httpCoordinatesClient{
		client: client,
	}, nil
}

type httpCoordinatesClient struct {
	client util_http.Client
}

func (h *httpCoordinatesClient) Coordinates() (coordinates.Coordinates, error) {
	result := coordinates.Coordinates{}
	req, err := http.NewRequest("GET", "/coordinates", nil)
	if err != nil {
		return result, errors.Wrap(err, "could not construct the request")
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return result, errors.Wrap(err, "could not execute the request")
	}
	if resp.StatusCode != 200 {
		return result, errors.Errorf("unexpected status code %d. Expected 200", resp.StatusCode)
	}

	coordinatesBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, errors.Wrap(err, "could not read a body of the request")
	}
	if err := json.Unmarshal(coordinatesBytes, &result); err != nil {
		return result, errors.Wrap(err, "could not unmarshal bytes to component")
	}
	return result, nil
}
