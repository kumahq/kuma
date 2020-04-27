package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	core_ca "github.com/Kong/kuma/pkg/core/ca"
	core_managers "github.com/Kong/kuma/pkg/core/managers/apis/mesh"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	k8s_resources "github.com/Kong/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"

	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MeshReconciler reconciles a Mesh object
type MeshReconciler struct {
	kube_client.Client
	Reader kube_client.Reader
	Log    logr.Logger

	Scheme          *kube_runtime.Scheme
	Converter       k8s_resources.Converter
	CaManagers      core_ca.Managers
	SystemNamespace string
	ResourceManager manager.ResourceManager
}

func (r *MeshReconciler) Reconcile(req kube_ctrl.Request) (kube_ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("mesh", req.NamespacedName)

	// Fetch the Mesh instance
	mesh := &mesh_k8s.Mesh{}
	if err := r.Get(ctx, req.NamespacedName, mesh); err != nil {
		if kube_apierrs.IsNotFound(err) {
			err := r.ResourceManager.Delete(ctx, &mesh_core.MeshResource{}, store.DeleteByKey(req.Name, req.Name))
			return kube_ctrl.Result{}, err
		}
		log.Error(err, "unable to fetch Mesh")
		return kube_ctrl.Result{}, err
	}

	meshResource := &mesh_core.MeshResource{}
	if err := r.Converter.ToCoreResource(mesh, meshResource); err != nil {
		log.Error(err, "unable to convert Mesh k8s object into core model")
		return kube_ctrl.Result{}, err
	}

	// Ensure CA Managers are created
	if err := core_managers.EnsureEnabledCA(ctx, r.CaManagers, meshResource, meshResource.Meta.GetName()); err != nil {
		log.Error(err, "unable to ensure that mesh CAs are created")
		return kube_ctrl.Result{}, err
	}

	return kube_ctrl.Result{}, nil
}

func (r *MeshReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	if err := kube_core.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", kube_core.SchemeGroupVersion)
	}
	if err := mesh_k8s.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", mesh_k8s.GroupVersion)
	}
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&mesh_k8s.Mesh{}).
		Complete(r)
}
