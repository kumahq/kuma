package resources

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type ZoneIngressInspectClient interface {
	InspectConfigDump(ctx context.Context, name string) ([]byte, error)
}

func NewZoneIngressInspectClient(client util_http.Client) ZoneIngressInspectClient {
	return &httpZoneIngressInspectClient{
		Client: client,
	}
}

type httpZoneIngressInspectClient struct {
	Client util_http.Client
}

var _ ZoneIngressInspectClient = &httpZoneIngressInspectClient{}

func (h *httpZoneIngressInspectClient) InspectConfigDump(ctx context.Context, name string) ([]byte, error) {
	resUrl, err := url.Parse(fmt.Sprintf("/zoneingresses/%s/xds", name))
	if err != nil {
		return nil, errors.Wrap(err, "could not construct the url")
	}
	req, err := http.NewRequest("GET", resUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	statusCode, b, err := doRequest(h.Client, ctx, req)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, errors.Errorf("(%d): %s", statusCode, string(b))
	}
	return b, nil
}
