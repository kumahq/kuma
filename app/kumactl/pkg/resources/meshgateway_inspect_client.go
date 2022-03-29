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
	InspectPolicies(ctx context.Context, mesh, name string) (api_server_types.GatewayInspectResult, error)
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

func (h *httpMeshGatewayInspectClient) InspectPolicies(ctx context.Context, mesh, name string) (api_server_types.GatewayInspectResult, error) {
	resUrl, err := url.Parse(fmt.Sprintf("/meshes/%s/meshgateways/%s/policies", mesh, name))
	if err != nil {
		return api_server_types.GatewayInspectResult{}, errors.Wrap(err, "could not construct the url")
	}
	req, err := http.NewRequest("GET", resUrl.String(), nil)
	if err != nil {
		return api_server_types.GatewayInspectResult{}, err
	}
	statusCode, b, err := doRequest(h.Client, ctx, req)
	if err != nil {
		return api_server_types.GatewayInspectResult{}, err
	}
	if statusCode != 200 {
		return api_server_types.GatewayInspectResult{}, errors.Errorf("(%d): %s", statusCode, string(b))
	}
	response := &api_server_types.GatewayInspectResult{}
	if err := json.Unmarshal(b, &response); err != nil {
		return api_server_types.GatewayInspectResult{}, err
	}
	return *response, nil
}
