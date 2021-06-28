package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	kumactl_client "github.com/kumahq/kuma/app/kumactl/pkg/client"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/resources/remote"
	kuma_http "github.com/kumahq/kuma/pkg/util/http"
)

type ServiceOverviewClient interface {
	List(ctx context.Context, mesh string) (*mesh.ServiceOverviewResourceList, error)
}

func NewServiceOverviewClient(coordinates *config_proto.ControlPlaneCoordinates_ApiServer) (ServiceOverviewClient, error) {
	client, err := kumactl_client.ApiServerClient(coordinates)
	if err != nil {
		return nil, err
	}
	return &httpServiceOverviewClient{
		Client: client,
	}, nil
}

type httpServiceOverviewClient struct {
	Client kuma_http.Client
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
		return nil, errors.Errorf("(%d): %s", statusCode, string(b))
	}
	overviews := &mesh.ServiceOverviewResourceList{}
	if err := remote.UnmarshalList(b, overviews); err != nil {
		return nil, err
	}
	return overviews, nil
}
