package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type PolicyInspectClient interface {
	Inspect(ctx context.Context, policyDesc core_model.ResourceTypeDescriptor, mesh, name string) (*api_server_types.PolicyInspectEntryList, error)
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

func (h *httpPolicyInspectClient) Inspect(ctx context.Context, policyDesc core_model.ResourceTypeDescriptor, mesh, name string) (*api_server_types.PolicyInspectEntryList, error) {
	resUrl, err := url.Parse(fmt.Sprintf("/meshes/%s/%s/%s/dataplanes", mesh, policyDesc.WsPath, name))
	if err != nil {
		return nil, errors.Wrap(err, "could not construct the url")
	}
	req, err := http.NewRequest("GET", resUrl.String(), nil)
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
