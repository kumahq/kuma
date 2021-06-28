package resources

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	kumactl_client "github.com/kumahq/kuma/app/kumactl/pkg/client"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/remote"
	kuma_http "github.com/kumahq/kuma/pkg/util/http"
)

type ZoneOverviewClient interface {
	List(ctx context.Context) (*system.ZoneOverviewResourceList, error)
}

func NewZoneOverviewClient(coordinates *config_proto.ControlPlaneCoordinates_ApiServer) (ZoneOverviewClient, error) {
	client, err := kumactl_client.ApiServerClient(coordinates)
	if err != nil {
		return nil, err
	}
	return &httpZoneOverviewClient{
		Client: client,
	}, nil
}

type httpZoneOverviewClient struct {
	Client kuma_http.Client
}

func (d *httpZoneOverviewClient) List(ctx context.Context) (*system.ZoneOverviewResourceList, error) {
	req, err := http.NewRequest("GET", "/zones+insights", nil)
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
	overviews := system.ZoneOverviewResourceList{}
	if err := remote.UnmarshalList(b, &overviews); err != nil {
		return nil, err
	}
	return &overviews, nil
}
