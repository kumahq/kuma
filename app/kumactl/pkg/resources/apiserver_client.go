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

type ApiServerClientFn struct {
	GetVersionFn func(ctx context.Context) (*types.IndexResponse, error)
	GetUserFn    func(ctx context.Context) (*api_server.WhoamiResponse, error)
}

func (fn ApiServerClientFn) GetVersion(ctx context.Context) (*types.IndexResponse, error) {
	return fn.GetVersionFn(ctx)
}

func (fn ApiServerClientFn) GetUser(ctx context.Context) (*api_server.WhoamiResponse, error) {
	return fn.GetUserFn(ctx)
}

func NewAPIServerClient(client util_http.Client) ApiServerClient {
	return ApiServerClientFn{
		GetVersionFn: func(ctx context.Context) (*types.IndexResponse, error) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				return nil, err
			}
			statusCode, b, err := doRequest(client, ctx, req)
			if err != nil {
				return nil, err
			}
			if statusCode != 200 {
				return nil, fmt.Errorf("(%d): %s", statusCode, string(b))
			}
			version := types.IndexResponse{}
			if err := json.Unmarshal(b, &version); err != nil {
				return nil, err
			}
			return &version, nil
		},
		GetUserFn: func(ctx context.Context) (*api_server.WhoamiResponse, error) {
			req, err := http.NewRequest("GET", "/who-am-i", nil)
			if err != nil {
				return nil, err
			}
			statusCode, b, err := doRequest(client, ctx, req)
			if err != nil {
				return nil, err
			}
			if statusCode != 200 {
				return nil, fmt.Errorf("(%d): %s", statusCode, string(b))
			}
			user := api_server.WhoamiResponse{}
			if err := json.Unmarshal(b, &user); err != nil {
				return nil, err
			}
			return &user, nil
		},
	}
}

func NewStaticApiServiceClient(version types.IndexResponse, user api_server.WhoamiResponse) ApiServerClient {
	return ApiServerClientFn{
		GetVersionFn: func(ctx context.Context) (*types.IndexResponse, error) {
			return &version, nil
		},
		GetUserFn: func(ctx context.Context) (*api_server.WhoamiResponse, error) {
			return &user, nil
		},
	}
}
