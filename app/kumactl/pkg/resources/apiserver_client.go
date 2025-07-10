package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kumahq/kuma/api/openapi/types"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type ApiServerClient interface {
	GetVersion(ctx context.Context) (*types.IndexResponse, error)
}

type ApiServerClientFn func(ctx context.Context) (*types.IndexResponse, error)

func (fn ApiServerClientFn) GetVersion(ctx context.Context) (*types.IndexResponse, error) {
	return fn(ctx)
}

func NewAPIServerClient(client util_http.Client) ApiServerClient {
	return ApiServerClientFn(func(ctx context.Context) (*types.IndexResponse, error) {
		req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
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
	})
}

func NewStaticApiServiceClient(response types.IndexResponse) ApiServerClient {
	return ApiServerClientFn(func(ctx context.Context) (*types.IndexResponse, error) {
		return &response, nil
	})
}
