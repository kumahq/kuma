package resources

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	kumactl_client "github.com/kumahq/kuma/app/kumactl/pkg/client"
	"github.com/kumahq/kuma/pkg/api-server/types"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	kuma_http "github.com/kumahq/kuma/pkg/util/http"
)

type ApiServerClient interface {
	GetVersion(ctx context.Context) (*types.IndexResponse, error)
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

func (d *httpApiServerClient) GetVersion(ctx context.Context) (*types.IndexResponse, error) {
	req, err := http.NewRequest("GET", "/", nil)
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
	version := types.IndexResponse{}
	if err := json.Unmarshal(b, &version); err != nil {
		return nil, err
	}
	return &version, nil
}
