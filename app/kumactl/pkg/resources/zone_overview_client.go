package resources

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	kumactl_client "github.com/kumahq/kuma/app/kumactl/pkg/client"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/rest/errors/types"
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
	statusCode, b, err := d.doRequest(ctx, req)
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

func (d *httpZoneOverviewClient) doRequest(ctx context.Context, req *http.Request) (int, []byte, error) {
	resp, err := d.Client.Do(req.WithContext(ctx))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	if resp.StatusCode/100 >= 4 {
		kumaErr := types.Error{}
		if err := json.Unmarshal(b, &kumaErr); err == nil {
			if kumaErr.Title != "" && kumaErr.Details != "" {
				return resp.StatusCode, b, &kumaErr
			}
		}
	}
	return resp.StatusCode, b, nil
}
