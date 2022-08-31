package resources

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type ZoneOverviewClient interface {
	List(ctx context.Context) (*system.ZoneOverviewResourceList, error)
}

func NewZoneOverviewClient(client util_http.Client) ZoneOverviewClient {
	return &httpZoneOverviewClient{
		Client: client,
	}
}

type httpZoneOverviewClient struct {
	Client util_http.Client
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
	if err := rest.JSON.UnmarshalListToCore(b, &overviews); err != nil {
		return nil, err
	}
	return &overviews, nil
}
