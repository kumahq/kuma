package k8s

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/discovery/k8s/controllers"

	kube_ctrl "sigs.k8s.io/controller-runtime"
)

func NewDiscoverySource(mgr kube_ctrl.Manager) (core_discovery.DiscoverySource, error) {
	source := &controllers.PodReconciler{
		Client: mgr.GetClient(),
		Log:    core.Log.WithName("controllers").WithName("Pod"),
	}
	if err := source.SetupWithManager(mgr); err != nil {
		return nil, err
	}
	return source, nil
}
