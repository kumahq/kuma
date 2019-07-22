package resources

import (
	"net/http"

	konvoyctl_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/k8s"
	konvoy_rest "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server/definitions"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	remote_resources "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/remote"
	"github.com/pkg/errors"
)

func NewResourceStore(controlPlane *config_proto.ControlPlane) (core_store.ResourceStore, error) {
	switch coordinates := controlPlane.Coordinates.Type.(type) {
	case *config_proto.ControlPlaneCoordinates_Kubernetes_:
		kubeConfig, err := konvoyctl_k8s.GetKubeConfig(coordinates.Kubernetes.Kubeconfig, coordinates.Kubernetes.Context, coordinates.Kubernetes.Namespace)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to load `kubectl` config: kubeconfig=%q context=%q namespace=%q",
				coordinates.Kubernetes.Kubeconfig, coordinates.Kubernetes.Context, coordinates.Kubernetes.Namespace)
		}
		// create a Transport that proxies requests to the Control Plane through kube-apiserver
		t, err := kubeConfig.NewServiceProxyTransport(coordinates.Kubernetes.Namespace, "konvoy-control-plane:http-api-server")
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create Kubernetes proxy transport")
		}
		return remote_resources.NewStore(http.Client{Transport: t}, konvoy_rest.AllApis()), nil
	default:
		return nil, errors.Errorf("Control Plane has coordinates that are not supported yet: %s", controlPlane.Coordinates.Type)
	}
}
