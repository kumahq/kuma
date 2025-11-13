package tokens

import (
	"time"

	"github.com/kumahq/kuma/v2/pkg/tokens"
	"github.com/kumahq/kuma/v2/pkg/tokens/builtin/server/types"
	util_http "github.com/kumahq/kuma/v2/pkg/util/http"
)

func NewZoneTokenClient(client util_http.Client) ZoneTokenClient {
	return &httpZoneTokenClient{
		client: tokens.NewTokenClient(client, "zone"),
	}
}

type ZoneTokenClient interface {
	Generate(zone string, scope []string, validFor time.Duration) (string, error)
}

type httpZoneTokenClient struct {
	client tokens.TokenClient
}

var _ ZoneTokenClient = &httpZoneTokenClient{}

func (h *httpZoneTokenClient) Generate(zone string, scope []string, validFor time.Duration) (string, error) {
	tokenReq := &types.ZoneTokenRequest{
		Zone:     zone,
		Scope:    scope,
		ValidFor: validFor.String(),
	}
	return h.client.Generate(tokenReq)
}
