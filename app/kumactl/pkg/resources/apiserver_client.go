package resources

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/api-server/types"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type ApiServerClient interface {
	GetVersion(ctx context.Context) (*types.IndexResponse, error)
}

func NewAPIServerClient(client util_http.Client) ApiServerClient {
	return &httpApiServerClient{
		Client: client,
	}
}

type httpApiServerClient struct {
	Client util_http.Client
}

func (d *httpApiServerClient) GetVersion(ctx context.Context) (*types.IndexResponse, error) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		return nil, err
	}
	statusCode, b, err := doRequest(d.Client, ctx, req)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, errors.Errorf("(%d): %s", statusCode, string(b))
	}
	version := types.IndexResponse{}
	if err := json.Unmarshal(b, &version); err != nil {
		return nil, err
	}
	return &version, nil
}
