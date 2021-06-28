package resources

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"

	kumactl_client "github.com/kumahq/kuma/app/kumactl/pkg/client"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/remote"
	kuma_http "github.com/kumahq/kuma/pkg/util/http"
)

type ZoneIngressOverviewClient interface {
	List(ctx context.Context) (*mesh.ZoneIngressOverviewResourceList, error)
}

func NewZoneIngressOverviewClient(coordinates *config_proto.ControlPlaneCoordinates_ApiServer) (ZoneIngressOverviewClient, error) {
	client, err := kumactl_client.ApiServerClient(coordinates)
	if err != nil {
		return nil, err
	}
	return &httpZoneIngressOverviewClient{
		Client: client,
	}, nil
}

type httpZoneIngressOverviewClient struct {
	Client kuma_http.Client
}

func (d *httpZoneIngressOverviewClient) List(ctx context.Context) (*mesh.ZoneIngressOverviewResourceList, error) {
	req, err := http.NewRequest("GET", "/zoneingresses+insights", nil)
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
	overviews := mesh.ZoneIngressOverviewResourceList{}
	if err := remote.UnmarshalList(b, &overviews); err != nil {
		return nil, err
	}
	return &overviews, nil
}
