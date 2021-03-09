package resources

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	kumactl_client "github.com/kumahq/kuma/app/kumactl/pkg/client"
	"github.com/kumahq/kuma/pkg/api-server/types"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	error_types "github.com/kumahq/kuma/pkg/core/rest/errors/types"
	kuma_http "github.com/kumahq/kuma/pkg/util/http"
)

type ApiServerClient interface {
	GetVersion() (*types.IndexResponse, error)
}

func NewAPIServerClient(coordinates *config_proto.ControlPlaneCoordinates_ApiServer) (ApiServerClient, error) {
	client, err := kumactl_client.ApiServerClient(coordinates)
	if err != nil {
		return nil, err
	}
	return &httpApiServerClient{
		Client: client,
	}, nil
}

type httpApiServerClient struct {
	Client kuma_http.Client
}

func (d *httpApiServerClient) GetVersion() (*types.IndexResponse, error) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		return nil, err
	}
	statusCode, b, err := d.doRequest(req)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, errors.Errorf("(%d): %s", statusCode, string(b))
	}
	version := types.IndexResponse{}
	if err := json.Unmarshal(b, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

func (d *httpApiServerClient) doRequest(req *http.Request) (int, []byte, error) {
	resp, err := d.Client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	if resp.StatusCode/100 >= 4 {
		kumaErr := error_types.Error{}
		if err := json.Unmarshal(b, &kumaErr); err == nil {
			if kumaErr.Title != "" && kumaErr.Details != "" {
				return resp.StatusCode, b, &kumaErr
			}
		}
	}
	return resp.StatusCode, b, nil
}
