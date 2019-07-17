package resources

import (
	konvoyctl_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/k8s"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	k8s_resources "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s"
	"github.com/pkg/errors"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
)

func NewResourceStore(controlPlane *config_proto.ControlPlane) (core_store.ResourceStore, error) {
	switch coordinates := controlPlane.Coordinates.Type.(type) {
	case *config_proto.ControlPlaneCoordinates_Kubernetes_:
		kubeConfig, err := konvoyctl_k8s.GetKubeConfig(coordinates.Kubernetes.Kubeconfig, coordinates.Kubernetes.Context, coordinates.Kubernetes.Namespace)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to load `kubectl` config: kubeconfig=%q context=%q namespace=%q",
				coordinates.Kubernetes.Kubeconfig, coordinates.Kubernetes.Context, coordinates.Kubernetes.Namespace)
		}
		kubeClient, err := kubeConfig.NewClient()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create Kubernetes client")
		}
		return k8s_resources.NewStore(kubeClient.(kube_client.Client))
	default:
		return nil, errors.Errorf("Control Plane has coordinates that are not supported yet: %s", controlPlane.Coordinates.Type)
	}
}
