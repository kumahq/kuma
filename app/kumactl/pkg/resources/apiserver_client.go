package resources

import (
	"context"
	"encoding/json"
	"fmt"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	"net/http"

	"github.com/kumahq/kuma/api/openapi/types"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type ApiServerClient interface {
	GetVersion(ctx context.Context) (*types.IndexResponse, error)
	GetUser(ctx context.Context) (*api_server.WhoamiResponse, error)
}

type apiServerClient struct {
	client util_http.Client
}

func NewAPIServerClient(client util_http.Client) ApiServerClient {
	return &apiServerClient{
		client: client,
	}
}

func (c *apiServerClient) GetVersion(ctx context.Context) (*types.IndexResponse, error) {
	var version types.IndexResponse
	if err := c.getJSON(ctx, "/", &version); err != nil {
		return nil, err
	}
	return &version, nil
}

func (c *apiServerClient) GetUser(ctx context.Context) (*api_server.WhoamiResponse, error) {
	var user api_server.WhoamiResponse
	if err := c.getJSON(ctx, "/who-am-i", &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *apiServerClient) getJSON(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return err
	}
	statusCode, body, err := doRequest(c.client, ctx, req)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("(%d): %s", statusCode, string(body))
	}
	if err := json.Unmarshal(body, out); err != nil {
		return err
	}
	return nil
}

func NewStaticApiServiceClient(version types.IndexResponse, user api_server.WhoamiResponse) ApiServerClient {
	return &staticApiServerClient{
		version: version,
		user:    user,
	}
}

type staticApiServerClient struct {
	version types.IndexResponse
	user    api_server.WhoamiResponse
}

func (c *staticApiServerClient) GetVersion(ctx context.Context) (*types.IndexResponse, error) {
	return &c.version, nil
}

func (c *staticApiServerClient) GetUser(ctx context.Context) (*api_server.WhoamiResponse, error) {
	return &c.user, nil
}
