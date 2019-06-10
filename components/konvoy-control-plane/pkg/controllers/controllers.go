package controllers

import (
	model_controllers "github.com/Kong/konvoy/components/konvoy-control-plane/model/controllers"
	util_manager "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/manager"
	ctrl "sigs.k8s.io/controller-runtime"
)

func SetupWithManager(mgr ctrl.Manager) error {
	return util_manager.SetupWithManager(
		mgr,
		&model_controllers.ProxyTemplateReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("ProxyTemplate"),
		},
		&model_controllers.ProxyReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("Proxy"),
		},
	)
}
