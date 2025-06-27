package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type MeshGatewayInspectClient interface {
	InspectDataplanes(ctx context.Context, mesh, name string) (api_server_types.GatewayDataplanesInspectEntryList, error)
}

func NewMeshGatewayInspectClient(client util_http.Client) MeshGatewayInspectClient {
	return &httpMeshGatewayInspectClient{
		Client: client,
	}
}

type httpMeshGatewayInspectClient struct {
	Client util_http.Client
}

var _ MeshGatewayInspectClient = &httpMeshGatewayInspectClient{}

func (h *httpMeshGatewayInspectClient) InspectDataplanes(ctx context.Context, mesh, name string) (api_server_types.GatewayDataplanesInspectEntryList, error) {
	resUrl, err := url.Parse(fmt.Sprintf("/meshes/%s/meshgateways/%s/dataplanes", mesh, name))
	if err != nil {
		return api_server_types.GatewayDataplanesInspectEntryList{}, errors.Wrap(err, "could not construct the url")
	}
	req, err := http.NewRequest("GET", resUrl.String(), http.NoBody)
	if err != nil {
		return api_server_types.GatewayDataplanesInspectEntryList{}, err
	}
	statusCode, b, err := doRequest(h.Client, ctx, req)
	if err != nil {
		return api_server_types.GatewayDataplanesInspectEntryList{}, err
	}
	if statusCode != 200 {
		return api_server_types.GatewayDataplanesInspectEntryList{}, errors.Errorf("(%d): %s", statusCode, string(b))
	}
	response := &api_server_types.GatewayDataplanesInspectEntryList{}
	if err := json.Unmarshal(b, &response); err != nil {
		return api_server_types.GatewayDataplanesInspectEntryList{}, err
	}
	return *response, nil
}
