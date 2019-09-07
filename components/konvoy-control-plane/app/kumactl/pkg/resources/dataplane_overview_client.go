package resources

import (
	"context"
	"fmt"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/kumactl/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/remote"
	konvoy_http "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/http"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

type DataplaneOverviewClient interface {
	List(ctx context.Context, meshName string, tags map[string]string) (*mesh.DataplaneOverviewResourceList, error)
}

func NewDataplaneOverviewClient(coordinates *config_proto.ControlPlaneCoordinates_ApiServer) (DataplaneOverviewClient, error) {
	client, err := apiServerClient(coordinates.Url)
	if err != nil {
		return nil, err
	}
	return &httpDataplaneOverviewClient{
		Client: client,
	}, nil
}

type httpDataplaneOverviewClient struct {
	Client konvoy_http.Client
}

func (d *httpDataplaneOverviewClient) List(ctx context.Context, meshName string, tags map[string]string) (*mesh.DataplaneOverviewResourceList, error) {
	resUrl, err := constructUrl(meshName, tags)
	if err != nil {
		return nil, errors.Wrap(err, "could not construct the url")
	}
	req, err := http.NewRequest("GET", resUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	statusCode, b, err := d.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, errors.Errorf("(%d): %s", statusCode, string(b))
	}
	overviews := mesh.DataplaneOverviewResourceList{}
	if err := remote.UnmarshalList(b, &overviews); err != nil {
		return nil, err
	}
	return &overviews, nil
}

func constructUrl(meshName string, tags map[string]string) (*url.URL, error) {
	result, err := url.Parse(fmt.Sprintf("/meshes/%s/dataplanes+insights", meshName))
	if err != nil {
		return nil, err
	}
	query := result.Query()
	for tag, value := range tags {
		query.Add("tag", fmt.Sprintf("%s:%s", tag, value))
	}
	result.RawQuery = query.Encode()
	return result, err
}

func (d *httpDataplaneOverviewClient) doRequest(ctx context.Context, req *http.Request) (int, []byte, error) {
	resp, err := d.Client.Do(req.WithContext(ctx))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, b, nil
}
