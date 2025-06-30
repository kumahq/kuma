package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/openapi/types"
	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type PolicyInspectClient interface {
	Inspect(ctx context.Context, policyDesc core_model.ResourceTypeDescriptor, mesh, name string) (*api_server_types.PolicyInspectEntryList, error)
	DataplanesForPolicy(ctx context.Context, desc core_model.ResourceTypeDescriptor, mesh string, name string) (types.InspectDataplanesForPolicyResponse, error)
}

func NewPolicyInspectClient(client util_http.Client) PolicyInspectClient {
	return &httpPolicyInspectClient{
		Client: client,
	}
}

var _ PolicyInspectClient = &httpPolicyInspectClient{}

type httpPolicyInspectClient struct {
	Client util_http.Client
}

func (h *httpPolicyInspectClient) DataplanesForPolicy(ctx context.Context, policyDesc core_model.ResourceTypeDescriptor, mesh string, name string) (types.InspectDataplanesForPolicyResponse, error) {
	resUrl, err := url.Parse(fmt.Sprintf("/meshes/%s/%s/%s/-resources/dataplanes", mesh, policyDesc.WsPath, name))
	if err != nil {
		return types.InspectDataplanesForPolicyResponse{}, errors.Wrap(err, "could not construct the url")
	}
	req, err := http.NewRequest("GET", resUrl.String(), http.NoBody)
	if err != nil {
		return types.InspectDataplanesForPolicyResponse{}, err
	}
	statusCode, b, err := doRequest(h.Client, ctx, req)
	if err != nil {
		return types.InspectDataplanesForPolicyResponse{}, err
	}
	if statusCode != 200 {
		return types.InspectDataplanesForPolicyResponse{}, errors.Errorf("(%d): %s", statusCode, string(b))
	}
	entryList := types.InspectDataplanesForPolicyResponse{}
	if err := json.Unmarshal(b, &entryList); err != nil {
		return types.InspectDataplanesForPolicyResponse{}, err
	}
	return entryList, nil
}

func (h *httpPolicyInspectClient) Inspect(ctx context.Context, policyDesc core_model.ResourceTypeDescriptor, mesh, name string) (*api_server_types.PolicyInspectEntryList, error) {
	resUrl, err := url.Parse(fmt.Sprintf("/meshes/%s/%s/%s/dataplanes", mesh, policyDesc.WsPath, name))
	if err != nil {
		return nil, errors.Wrap(err, "could not construct the url")
	}
	req, err := http.NewRequest("GET", resUrl.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	statusCode, b, err := doRequest(h.Client, ctx, req)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, errors.Errorf("(%d): %s", statusCode, string(b))
	}
	entryList := &api_server_types.PolicyInspectEntryList{}
	if err := json.Unmarshal(b, entryList); err != nil {
		return nil, err
	}
	return entryList, nil
}
