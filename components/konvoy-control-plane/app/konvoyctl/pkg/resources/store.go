package resources

import (
	"net/http"
	"net/url"
	"time"

	konvoyctl_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/k8s"
	konvoy_rest "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server/definitions"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	remote_resources "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/remote"
	util_http "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/http"
	"github.com/pkg/errors"
)

var (
	// overridable by unit tests
	getKubeConfig = konvoyctl_k8s.GetKubeConfig
)

const (
	// Time limit for requests to the Control Plane API Server.
	Timeout = 60 * time.Second
)

func NewResourceStore(controlPlane *config_proto.ControlPlane) (core_store.ResourceStore, error) {
	baseURL, err := url.Parse(controlPlane.GetCoordinates().ApiServer.Url)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse API Server URL")
	}
	client := util_http.ClientWithBaseURL(newClient(), baseURL)
	return remote_resources.NewStore(client, konvoy_rest.AllApis()), nil
}

func newClient() *http.Client {
	return &http.Client{
		Timeout: Timeout,
	}
}
