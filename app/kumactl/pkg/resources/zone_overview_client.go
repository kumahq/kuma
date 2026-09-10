package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	zone_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zone/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	util_http "github.com/kumahq/kuma/v3/pkg/util/http"
)

type ZoneOverviewClient interface {
	List(ctx context.Context) (*zone_api.ZoneOverviewResourceList, error)
}

func NewZoneOverviewClient(client util_http.Client) ZoneOverviewClient {
	return &httpZoneOverviewClient{
		Client: client,
	}
}

type httpZoneOverviewClient struct {
	Client util_http.Client
}

func (d *httpZoneOverviewClient) List(ctx context.Context) (*zone_api.ZoneOverviewResourceList, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/%s/_overview", zone_api.ZoneResourceTypeDescriptor.WsPath), http.NoBody)
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
	overviews := zone_api.ZoneOverviewResourceList{}
	if err := rest.JSON.UnmarshalListToCore(b, &overviews); err != nil {
		return nil, err
	}
	return &overviews, nil
}
