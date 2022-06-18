package tokens

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	error_types "github.com/kumahq/kuma/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/pkg/tokens/builtin/server/types"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

func NewZoneIngressTokenClient(client util_http.Client) ZoneIngressTokenClient {
	return &httpZoneIngressTokenClient{
		client: client,
	}
}

type ZoneIngressTokenClient interface {
	Generate(zone string, validFor time.Duration) (string, error)
}

type httpZoneIngressTokenClient struct {
	client util_http.Client
}

var _ ZoneIngressTokenClient = &httpZoneIngressTokenClient{}

func (h *httpZoneIngressTokenClient) Generate(zone string, validFor time.Duration) (string, error) {
	tokenReq := &types.ZoneIngressTokenRequest{
		Zone:     zone,
		ValidFor: validFor.String(),
	}
	reqBytes, err := json.Marshal(tokenReq)
	if err != nil {
		return "", fmt.Errorf("could not marshal token request to json: %w", err)
	}
	req, err := http.NewRequest("POST", "/tokens/zone-ingress", bytes.NewReader(reqBytes))
	if err != nil {
		return "", fmt.Errorf("could not construct the request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not execute the request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read a body of the request: %w", err)
	}
	if resp.StatusCode != 200 {
		kumaErr := error_types.Error{}
		if err := json.Unmarshal(body, &kumaErr); err == nil {
			if kumaErr.Title != "" && kumaErr.Details != "" {
				return "", &kumaErr
			}
		}
		return "", fmt.Errorf("(%d): %s", resp.StatusCode, body)
	}
	return string(body), nil
}
