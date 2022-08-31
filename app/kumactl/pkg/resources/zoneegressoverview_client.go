package resources

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type ZoneEgressOverviewClient interface {
	List(ctx context.Context) (*mesh.ZoneEgressOverviewResourceList, error)
}

func NewZoneEgressOverviewClient(client util_http.Client) ZoneEgressOverviewClient {
	return &httpZoneEgressOverviewClient{
		Client: client,
	}
}

type httpZoneEgressOverviewClient struct {
	Client util_http.Client
}

func (d *httpZoneEgressOverviewClient) List(ctx context.Context) (*mesh.ZoneEgressOverviewResourceList, error) {
	req, err := http.NewRequest("GET", "/zoneegressoverviews", nil)
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
	overviews := mesh.ZoneEgressOverviewResourceList{}
	if err := rest.JSON.UnmarshalListToCore(b, &overviews); err != nil {
		return nil, err
	}
	return &overviews, nil
}
