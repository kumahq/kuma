package api

import util_http "github.com/kumahq/kuma/v3/pkg/util/http"

type AuthnPlugin interface {
	Validate(map[string]string) error
	DecorateClient(util_http.Client, map[string]string) (util_http.Client, error)
}
