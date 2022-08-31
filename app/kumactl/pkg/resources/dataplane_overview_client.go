package resources

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type DataplaneOverviewClient interface {
	List(ctx context.Context, meshName string, tags map[string]string, gateway bool, ingress bool) (*mesh.DataplaneOverviewResourceList, error)
}

func NewDataplaneOverviewClient(client util_http.Client) DataplaneOverviewClient {
	return &httpDataplaneOverviewClient{
		Client: client,
	}
}

type httpDataplaneOverviewClient struct {
	Client util_http.Client
}

func (d *httpDataplaneOverviewClient) List(ctx context.Context, meshName string, tags map[string]string, gateway bool, ingress bool) (*mesh.DataplaneOverviewResourceList, error) {
	resUrl, err := constructUrl(meshName, tags, gateway, ingress)
	if err != nil {
		return nil, errors.Wrap(err, "could not construct the url")
	}
	req, err := http.NewRequest("GET", resUrl.String(), nil)
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
	overviews := mesh.DataplaneOverviewResourceList{}
	if err := rest.JSON.UnmarshalListToCore(b, &overviews); err != nil {
		return nil, err
	}
	return &overviews, nil
}

func constructUrl(meshName string, tags map[string]string, gateway bool, ingress bool) (*url.URL, error) {
	result, err := url.Parse(fmt.Sprintf("/meshes/%s/dataplanes+insights", meshName))
	if err != nil {
		return nil, err
	}
	query := result.Query()
	if gateway {
		query.Add("gateway", fmt.Sprintf("%t", gateway))
	}
	if ingress {
		query.Add("ingress", fmt.Sprintf("%t", ingress))
	}
	for tag, value := range tags {
		query.Add("tag", fmt.Sprintf("%s:%s", tag, value))
	}
	result.RawQuery = query.Encode()
	return result, err
}
