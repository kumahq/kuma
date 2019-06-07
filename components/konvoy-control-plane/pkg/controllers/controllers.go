package controllers

import (
	model_controllers "github.com/Kong/konvoy/components/konvoy-control-plane/model/controllers"
	ctrl "sigs.k8s.io/controller-runtime"
)

func SetupWithManager(mgr ctrl.Manager) error {
	return (&model_controllers.ProxyReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Proxy"),
	}).SetupWithManager(mgr)
}
