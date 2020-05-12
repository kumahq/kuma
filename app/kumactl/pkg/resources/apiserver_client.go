package resources

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/api-server/types"
	kuma_http "github.com/Kong/kuma/pkg/util/http"
)

const (
	timeout = 10 * time.Second
)

type ApiServerClient interface {
	GetVersion() (types.IndexResponse, error)
}

func NewApiServerClient(address string) (ApiServerClient, error) {
	client, err := apiServerClient(address)
	if err != nil {
		return nil, err
	}
	return &httpApiServerClient{
		Client: client,
	}, nil
}

type httpApiServerClient struct {
	Client kuma_http.Client
}

func (h *httpApiServerClient) GetVersion() (types.IndexResponse, error) {
	result := types.IndexResponse{}
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		return result, errors.Wrap(err, "could not construct the request")
	}
	resp, err := h.Client.Do(req)
	if err != nil {
		return result, errors.Wrap(err, "could not execute the request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return result, errors.Errorf("unexpected status code %d", resp.StatusCode)
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
