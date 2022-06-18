package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	error_types "github.com/kumahq/kuma/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type UserTokenClient interface {
	Generate(name string, groups []string, validFor time.Duration) (string, error)
}

var _ UserTokenClient = &httpUserTokenClient{}

func NewHTTPUserTokenClient(client util_http.Client) UserTokenClient {
	return &httpUserTokenClient{
		client: client,
	}
}

type httpUserTokenClient struct {
	client util_http.Client
}

func (h *httpUserTokenClient) Generate(name string, groups []string, validFor time.Duration) (string, error) {
	tokenReq := &ws.UserTokenRequest{
		Name:     name,
		Groups:   groups,
		ValidFor: validFor.String(),
	}
	reqBytes, err := json.Marshal(tokenReq)
	if err != nil {
		return "", fmt.Errorf("could not marshal token request to json: %w", err)
	}
	req, err := http.NewRequest("POST", "/tokens/user", bytes.NewReader(reqBytes))
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
		var kumaErr error_types.Error
		if err := json.Unmarshal(body, &kumaErr); err == nil {
			if kumaErr.Title != "" && kumaErr.Details != "" {
				return "", &kumaErr
			}
		}
		return "", fmt.Errorf("(%d): %s", resp.StatusCode, body)
	}
	return string(body), nil
}
