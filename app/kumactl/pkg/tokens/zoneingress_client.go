package tokens

import (
	"time"

	"github.com/kumahq/kuma/pkg/tokens"
	"github.com/kumahq/kuma/pkg/tokens/builtin/server/types"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

func NewZoneIngressTokenClient(client util_http.Client) ZoneIngressTokenClient {
	return &httpZoneIngressTokenClient{
		client: tokens.NewTokenClient(client, "zone-ingress"),
	}
}

type ZoneIngressTokenClient interface {
	Generate(zone string, validFor time.Duration) (string, error)
}

type httpZoneIngressTokenClient struct {
	client tokens.TokenClient
}

var _ ZoneIngressTokenClient = &httpZoneIngressTokenClient{}

func (h *httpZoneIngressTokenClient) Generate(zone string, validFor time.Duration) (string, error) {
	tokenReq := &types.ZoneIngressTokenRequest{
		Zone:     zone,
		ValidFor: validFor.String(),
	}
	return h.client.Generate(tokenReq)
}
