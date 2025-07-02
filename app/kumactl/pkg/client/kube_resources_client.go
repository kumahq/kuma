package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type KubernetesResourcesClient interface {
	Get(ctx context.Context, descriptor model.ResourceTypeDescriptor, name, mesh string) (map[string]interface{}, error)
}

type HTTPKubernetesResourcesClient struct {
	client util_http.Client
	api    rest.Api
}

func NewHTTPKubernetesResourcesClient(client util_http.Client, defs []model.ResourceTypeDescriptor) KubernetesResourcesClient {
	mapping := make(map[model.ResourceType]rest.ResourceApi)
	for _, ws := range defs {
		mapping[ws.Name] = rest.NewResourceApi(ws.Scope, ws.WsPath)
	}
	return &HTTPKubernetesResourcesClient{
		client: client,
		api: &rest.ApiDescriptor{
			Resources: mapping,
		},
	}
}

func (k *HTTPKubernetesResourcesClient) Get(ctx context.Context, descriptor model.ResourceTypeDescriptor, name, mesh string) (map[string]interface{}, error) {
	api, err := k.api.GetResourceApi(descriptor.Name)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, api.Item(mesh, name)+"?format=k8s", http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := k.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("unexpected status code: %d %s", resp.StatusCode, b)
	}
	obj := map[string]interface{}{}
	if err := json.Unmarshal(b, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}
