package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	api_common "github.com/kumahq/kuma/v3/api/openapi/types/common"
	util_http "github.com/kumahq/kuma/v3/pkg/util/http"
)

type DataplaneInspectClient interface {
	InspectPolicies(ctx context.Context, mesh, name string) (api_common.PoliciesList, error)
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

func (h *httpDataplaneInspectClient) InspectPolicies(ctx context.Context, mesh, name string) (api_common.PoliciesList, error) {
	resUrl, err := url.Parse(fmt.Sprintf("/meshes/%s/dataplanes/%s/_policies", mesh, name))
	if err != nil {
		return api_common.PoliciesList{}, errors.Wrap(err, "could not construct the url")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resUrl.String(), http.NoBody)
	if err != nil {
		return api_common.PoliciesList{}, err
	}
	statusCode, b, err := doRequest(h.Client, ctx, req)
	if err != nil {
		return api_common.PoliciesList{}, err
	}
	if statusCode != 200 {
		return api_common.PoliciesList{}, errors.Errorf("(%d): %s", statusCode, string(b))
	}
	var response api_common.PoliciesList
	if err := json.Unmarshal(b, &response); err != nil {
		return api_common.PoliciesList{}, err
	}
	return response, nil
}
