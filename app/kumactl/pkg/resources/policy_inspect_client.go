package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v3/api/openapi/types"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	util_http "github.com/kumahq/kuma/v3/pkg/util/http"
)

type PolicyInspectClient interface {
	DataplanesForPolicy(ctx context.Context, desc core_model.ResourceTypeDescriptor, mesh, name string, size int, offset string) (types.InspectDataplanesForPolicyResponse, error)
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

func (h *httpPolicyInspectClient) DataplanesForPolicy(ctx context.Context, policyDesc core_model.ResourceTypeDescriptor, mesh, name string, size int, offset string) (types.InspectDataplanesForPolicyResponse, error) {
	resUrl, err := url.Parse(fmt.Sprintf("/meshes/%s/%s/%s/_resources/dataplanes", mesh, policyDesc.WsPath, name))
	if err != nil {
		return types.InspectDataplanesForPolicyResponse{}, errors.Wrap(err, "could not construct the url")
	}
	query := resUrl.Query()
	if size != 0 {
		query.Set("size", strconv.Itoa(size))
	}
	if offset != "" {
		query.Set("offset", offset)
	}
	resUrl.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resUrl.String(), http.NoBody)
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
