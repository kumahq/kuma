package tokens

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"

	error_types "github.com/kumahq/kuma/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/pkg/tokens/builtin/server/types"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

func NewZoneTokenClient(client util_http.Client) ZoneTokenClient {
	return &httpZoneTokenClient{
		client: client,
	}
}

type ZoneTokenClient interface {
	Generate(zone string, scope []string, validFor time.Duration) (string, error)
}

type httpZoneTokenClient struct {
	client util_http.Client
}

var _ ZoneTokenClient = &httpZoneTokenClient{}

func (h *httpZoneTokenClient) Generate(zone string, scope []string, validFor time.Duration) (string, error) {
	tokenReq := &types.ZoneTokenRequest{
		Zone:     zone,
		Scope:    scope,
		ValidFor: validFor.String(),
	}

	reqBytes, err := json.Marshal(tokenReq)
	if err != nil {
		return "", errors.Wrap(err, "could not marshal token request to json")
	}

	req, err := http.NewRequest(http.MethodPost, "/tokens/zone", bytes.NewReader(reqBytes))
	if err != nil {
		return "", errors.Wrap(err, "could not construct the request")
	}

	req.Header.Set("content-type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "could not execute the request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "could not read a body of the request")
	}

	if resp.StatusCode != http.StatusOK {
		var kumaErr error_types.Error
		if err := json.Unmarshal(body, &kumaErr); err == nil {
			if kumaErr.Title != "" && kumaErr.Details != "" {
				return "", &kumaErr
			}
		}
		return "", errors.Errorf("(%d): %s", resp.StatusCode, body)
	}

	return string(body), nil
}
