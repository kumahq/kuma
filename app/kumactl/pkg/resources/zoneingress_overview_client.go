package resources

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type ZoneIngressOverviewClient interface {
	List(ctx context.Context) (*mesh.ZoneIngressOverviewResourceList, error)
}

func NewZoneIngressOverviewClient(client util_http.Client) ZoneIngressOverviewClient {
	return &httpZoneIngressOverviewClient{
		Client: client,
	}
}

type httpZoneIngressOverviewClient struct {
	Client util_http.Client
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
	if err := rest.JSON.UnmarshalListToCore(b, &overviews); err != nil {
		return nil, err
	}
	return &overviews, nil
}
