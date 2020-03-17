package client

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/catalog"
	util_http "github.com/Kong/kuma/pkg/util/http"
)

const (
	timeout = 10 * time.Second
)

type CatalogClient interface {
	Catalog() (catalog.Catalog, error)
}

func NewCatalogClient(address string) (CatalogClient, error) {
	baseURL, err := url.Parse(address)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse API Server URL")
	}
	client := &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	return &httpCatalogClient{
		client: util_http.ClientWithBaseURL(client, baseURL),
	}, nil
}

type httpCatalogClient struct {
	client util_http.Client
}

func (h *httpCatalogClient) Catalog() (catalog.Catalog, error) {
	result := catalog.Catalog{}
	req, err := http.NewRequest("GET", "/catalog", nil)
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

	catalogBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, errors.Wrap(err, "could not read a body of the request")
	}
	if err := json.Unmarshal(catalogBytes, &result); err != nil {
		return result, errors.Wrap(err, "could not unmarshal bytes to component")
	}
	return result, nil
}
