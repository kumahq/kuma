package test

import (
	"context"

	kumactl_resources "github.com/kumahq/kuma/app/kumactl/pkg/resources"
	"github.com/kumahq/kuma/pkg/api-server/types"
	util_http "github.com/kumahq/kuma/pkg/util/http"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type MockAPIServerClient struct {
	Version types.IndexResponse
}

func (m *MockAPIServerClient) GetVersion(ctx context.Context) (*types.IndexResponse, error) {
	return &m.Version, nil
}

func GetMockNewAPIServerClient() func(util_http.Client) kumactl_resources.ApiServerClient {
	return func(util_http.Client) kumactl_resources.ApiServerClient {
		return &MockAPIServerClient{
			Version: types.IndexResponse{
				Hostname: "localhost",
				Tagline:  kuma_version.Product,
				Version:  kuma_version.Build.Version,
			},
		}
	}
}
