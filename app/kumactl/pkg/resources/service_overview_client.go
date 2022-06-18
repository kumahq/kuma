package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/resources/remote"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type ServiceOverviewClient interface {
	List(ctx context.Context, mesh string) (*mesh.ServiceOverviewResourceList, error)
}

func NewServiceOverviewClient(client util_http.Client) ServiceOverviewClient {
	return &httpServiceOverviewClient{
		Client: client,
	}
}

type httpServiceOverviewClient struct {
	Client util_http.Client
}

func (d *httpServiceOverviewClient) List(ctx context.Context, meshName string) (*mesh.ServiceOverviewResourceList, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/meshes/%s/service-insights", meshName), nil)
	if err != nil {
		return nil, err
	}
	statusCode, b, err := doRequest(d.Client, ctx, req)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("(%d): %s", statusCode, string(b))
	}
	overviews := &mesh.ServiceOverviewResourceList{}
	if err := remote.UnmarshalList(b, overviews); err != nil {
		return nil, err
	}
	return overviews, nil
}
