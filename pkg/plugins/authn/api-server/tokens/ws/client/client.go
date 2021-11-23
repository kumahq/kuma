package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"

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
		return "", errors.Wrap(err, "could not marshal token request to json")
	}
	req, err := http.NewRequest("POST", "/tokens/user", bytes.NewReader(reqBytes))
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
	if resp.StatusCode != 200 {
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
