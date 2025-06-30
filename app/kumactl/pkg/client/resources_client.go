package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"

	api_types "github.com/kumahq/kuma/api/openapi/types"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type ResourcesListClient interface {
	List(ctx context.Context) (api_types.ResourceTypeDescriptionList, error)
}

var _ ResourcesListClient = &HTTPResourcesListClient{}

type HTTPResourcesListClient struct {
	client util_http.Client
}

func NewHTTPResourcesListClient(client util_http.Client) ResourcesListClient {
	return &HTTPResourcesListClient{
		client: client,
	}
}

func (h HTTPResourcesListClient) List(ctx context.Context) (api_types.ResourceTypeDescriptionList, error) {
	req, err := http.NewRequest("GET", "/_resources", http.NoBody)
	if err != nil {
		return api_types.ResourceTypeDescriptionList{}, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := h.client.Do(req.WithContext(ctx))
	if err != nil {
		return api_types.ResourceTypeDescriptionList{}, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return api_types.ResourceTypeDescriptionList{}, err
	}
	if resp.StatusCode != 200 {
		return api_types.ResourceTypeDescriptionList{}, errors.Errorf("unexpected status code: %d %s", resp.StatusCode, b)
	}
	list := api_types.ResourceTypeDescriptionList{}
	if err := json.Unmarshal(b, &list); err != nil {
		return api_types.ResourceTypeDescriptionList{}, err
	}
	return list, nil
}
