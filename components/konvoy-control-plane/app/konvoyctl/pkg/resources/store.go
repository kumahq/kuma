package resources

import (
	"net/http"
	"net/url"

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

func NewResourceStore(controlPlane *config_proto.ControlPlane) (core_store.ResourceStore, error) {
	switch coordinates := controlPlane.GetCoordinates().GetType().(type) {
	case *config_proto.ControlPlaneCoordinates_Kubernetes_:
		kubeConfig, err := getKubeConfig(coordinates.Kubernetes.Kubeconfig, coordinates.Kubernetes.Context, coordinates.Kubernetes.Namespace)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to load `kubectl` config: kubeconfig=%q context=%q namespace=%q",
				coordinates.Kubernetes.Kubeconfig, coordinates.Kubernetes.Context, coordinates.Kubernetes.Namespace)
		}
		// create a Transport that proxies requests to the Control Plane through kube-apiserver
		t, err := kubeConfig.NewServiceProxyTransport(coordinates.Kubernetes.Namespace, "konvoy-control-plane:http-api-server")
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create Kubernetes proxy transport")
		}
		return remote_resources.NewStore(&http.Client{Transport: t}, konvoy_rest.AllApis()), nil
	case *config_proto.ControlPlaneCoordinates_ApiServer_:
		baseURL, err := url.Parse(coordinates.ApiServer.Url)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to parse API Server URL")
		}
		client := util_http.ClientWithBaseURL(&http.Client{}, baseURL)
		return remote_resources.NewStore(client, konvoy_rest.AllApis()), nil
	default:
		return nil, errors.Errorf("Control Plane has coordinates that are not supported yet: %v", controlPlane.GetCoordinates().GetType())
	}
}
