package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/pkg/plugins/resources/remote"
	kuma_http "github.com/kumahq/kuma/pkg/util/http"
)

type ServiceOverviewClient interface {
	List(ctx context.Context, mesh string) (*mesh.ServiceOverviewResourceList, error)
}

func NewServiceOverviewClient(coordinates *config_proto.ControlPlaneCoordinates_ApiServer) (ServiceOverviewClient, error) {
	client, err := apiServerClient(coordinates)
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
	statusCode, b, err := d.doRequest(ctx, req)
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

func (d *httpServiceOverviewClient) doRequest(ctx context.Context, req *http.Request) (int, []byte, error) {
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
