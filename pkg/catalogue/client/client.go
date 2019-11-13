package client

import (
	"encoding/json"
	"github.com/Kong/kuma/pkg/catalogue"
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

type CatalogueClient interface {
	Catalogue() (catalogue.Catalogue, error)
}

func NewCatalogueClient(address string) (CatalogueClient, error) {
	baseURL, err := url.Parse(address)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse API Server URL")
	}
	client := &http.Client{
		Timeout: timeout,
	}
	return &httpCatalogueClient{
		client: util_http.ClientWithBaseURL(client, baseURL),
	}, nil
}

type httpCatalogueClient struct {
	client util_http.Client
}

func (h *httpCatalogueClient) Catalogue() (catalogue.Catalogue, error) {
	result := catalogue.Catalogue{}
	req, err := http.NewRequest("GET", "/catalogue", nil)
	if err != nil {
		return result, errors.Wrap(err, "could not construct the request")
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return result, errors.Wrap(err, "could not execute the request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return result, errors.Errorf("unexpected status code %d. Expected 200", resp.StatusCode)
	}

	catalogueBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, errors.Wrap(err, "could not read a body of the request")
	}
	if err := json.Unmarshal(catalogueBytes, &result); err != nil {
		return result, errors.Wrap(err, "could not unmarshal bytes to component")
	}
	return result, nil
}
