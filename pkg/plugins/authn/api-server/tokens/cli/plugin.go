package cli

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/app/kumactl/pkg/plugins"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

const (
	AuthType = "tokens"
	TokenKey = "token"
)

type TokenAuthnPlugin struct{}

var _ plugins.AuthnPlugin = &TokenAuthnPlugin{}

func (t *TokenAuthnPlugin) Validate(authConf map[string]string) error {
	if authConf[TokenKey] == "" {
		return errors.New("provide token=YOUR_TOKEN")
	}
	return nil
}

func (t *TokenAuthnPlugin) DecorateClient(delegate util_http.Client, authConf map[string]string) (util_http.Client, error) {
	return util_http.ClientFunc(func(req *http.Request) (*http.Response, error) {
		req.Header.Set("authorization", "Bearer "+authConf[TokenKey])
		return delegate.Do(req)
	}), nil
}
