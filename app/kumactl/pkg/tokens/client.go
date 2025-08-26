package tokens

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/tokens"
	"github.com/kumahq/kuma/pkg/tokens/builtin/server/types"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

func NewDataplaneTokenClient(client util_http.Client) DataplaneTokenClient {
	return &httpDataplaneTokenClient{
		client: tokens.NewTokenClient(client, "dataplane"),
	}
}

type DataplaneTokenClient interface {
	Generate(name, mesh string, tags map[string][]string, dpType string, validFor time.Duration) (string, error)
}

type httpDataplaneTokenClient struct {
	client tokens.TokenClient
}

var _ DataplaneTokenClient = &httpDataplaneTokenClient{}

func (h *httpDataplaneTokenClient) Generate(name, mesh string, tags map[string][]string, dpType string, validFor time.Duration) (string, error) {
	if validFor == 0 {
		return "", errors.Errorf("You must set a token validFor value")
	}
	tokenReq := &types.DataplaneTokenRequest{
		Name:     name,
		Mesh:     mesh,
		Tags:     tags,
		Type:     dpType,
		ValidFor: validFor.String(),
	}
	return h.client.Generate(tokenReq)
}
