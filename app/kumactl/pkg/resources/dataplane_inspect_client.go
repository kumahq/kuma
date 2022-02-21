package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type DataplaneInspectClient interface {
	InspectPolicies(ctx context.Context, mesh, name string) (*api_server_types.DataplaneInspectEntryList, error)
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

func (h *httpDataplaneInspectClient) InspectPolicies(ctx context.Context, mesh, name string) (*api_server_types.DataplaneInspectEntryList, error) {
	resUrl, err := url.Parse(fmt.Sprintf("/meshes/%s/dataplanes/%s/policies", mesh, name))
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
	receiver := &api_server_types.DataplaneInspectEntryListReceiver{
		NewResource: registry.Global().NewObject,
	}
	if err := json.Unmarshal(b, receiver); err != nil {
		return nil, err
	}
	return &receiver.DataplaneInspectEntryList, nil
}
