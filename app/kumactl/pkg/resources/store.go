package resources

import (
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	admin_rest "github.com/Kong/kuma/pkg/admin-server/definitions"
	kuma_rest "github.com/Kong/kuma/pkg/api-server/definitions"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	remote_resources "github.com/Kong/kuma/pkg/plugins/resources/remote"
	util_http "github.com/Kong/kuma/pkg/util/http"
)

func NewResourceStore(coordinates *config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
	client, err := apiServerClient(coordinates.Url)
	if err != nil {
		return nil, err
	}
	return remote_resources.NewStore(client, kuma_rest.AllApis()), nil
}

func NewAdminResourceStore(address string, config *config_proto.Context_AdminApiCredentials) (core_store.ResourceStore, error) {
	baseURL, err := url.Parse(address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse the server URL")
	}
	httpClient := &http.Client{
		Timeout:   Timeout,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	if baseURL.Scheme == "https" {
		if !config.HasClientCert() {
			return nil, errors.New("certificates has to be configured to use https destination")
		}
		if err := util_http.ConfigureTlsWithoutServerVerification(httpClient, config.ClientCert, config.ClientKey); err != nil {
			return nil, errors.Wrap(err, "could not configure tls")
		}
	}
	client := util_http.ClientWithBaseURL(httpClient, baseURL)

	return remote_resources.NewStore(client, admin_rest.AllApis()), nil
}
