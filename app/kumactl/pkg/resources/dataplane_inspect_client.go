package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type DataplaneInspectClient interface {
	InspectPolicies(ctx context.Context, mesh, name string) (api_server_types.DataplaneInspectResponse, error)
}

func NewDataplaneInspectClient(client util_http.Client) DataplaneInspectClient {
	return &httpDataplaneInspectClient{
		Client: client,
	}
}

type httpDataplaneInspectClient struct {
	Client util_http.Client
}

var _ DataplaneInspectClient = &httpDataplaneInspectClient{}

func (h *httpDataplaneInspectClient) InspectPolicies(ctx context.Context, mesh, name string) (api_server_types.DataplaneInspectResponse, error) {
	resUrl, err := url.Parse(fmt.Sprintf("/meshes/%s/dataplanes/%s/policies", mesh, name))
	if err != nil {
		return api_server_types.DataplaneInspectResponse{}, fmt.Errorf("could not construct the url: %w", err)
	}
	req, err := http.NewRequest("GET", resUrl.String(), nil)
	if err != nil {
		return api_server_types.DataplaneInspectResponse{}, err
	}
	statusCode, b, err := doRequest(h.Client, ctx, req)
	if err != nil {
		return api_server_types.DataplaneInspectResponse{}, err
	}
	if statusCode != 200 {
		return api_server_types.DataplaneInspectResponse{}, fmt.Errorf("(%d): %s", statusCode, string(b))
	}
	response := &api_server_types.DataplaneInspectResponse{}
	if err := json.Unmarshal(b, &response); err != nil {
		return api_server_types.DataplaneInspectResponse{}, err
	}
	return *response, nil
}
