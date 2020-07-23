/*
Copyright 2019 Kuma authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	meshv1alpha1 "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

// ProxyTemplateReconciler reconciles a ProxyTemplate object
type ProxyTemplateReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=kuma.io,resources=proxytemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kuma.io,resources=proxytemplates/status,verbs=get;update;patch

func (r *ProxyTemplateReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("proxytemplate", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *ProxyTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := meshv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&meshv1alpha1.ProxyTemplate{}).
		Complete(r)
}
