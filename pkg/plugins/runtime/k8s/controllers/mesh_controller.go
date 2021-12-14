package controllers

import (
	"context"

	"github.com/go-logr/logr"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_managers "github.com/kumahq/kuma/pkg/core/managers/apis/mesh"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

// MeshReconciler reconciles a Mesh object
type MeshReconciler struct {
	kube_client.Client
	Log logr.Logger

	Scheme          *kube_runtime.Scheme
	Converter       k8s_common.Converter
	CaManagers      core_ca.Managers
	SystemNamespace string
	ResourceManager manager.ResourceManager
}

func (r *MeshReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	log := r.Log.WithValues("mesh", req.NamespacedName)

	// Fetch the Mesh instance
	mesh := &mesh_k8s.Mesh{}
	if err := r.Get(ctx, req.NamespacedName, mesh); err != nil {
		if kube_apierrs.IsNotFound(err) {
			// Force delete associated resources. It will return an error ErrorResourceNotFound because Mesh was already deleted but we still need to cleanup resources
			// Remove this part after https://github.com/kumahq/kuma/issues/1137 is implemented.
			err := r.ResourceManager.Delete(ctx, core_mesh.NewMeshResource(), store.DeleteByKey(req.Name, req.Name))
			if err == nil || store.IsResourceNotFound(err) {
				return kube_ctrl.Result{}, nil
			}
			return kube_ctrl.Result{}, err
		}
		log.Error(err, "unable to fetch Mesh")
		return kube_ctrl.Result{}, err
	}

	meshResource := core_mesh.NewMeshResource()
	if err := r.Converter.ToCoreResource(mesh, meshResource); err != nil {
		log.Error(err, "unable to convert Mesh k8s object into core model")
		return kube_ctrl.Result{}, err
	}

	// Ensure CA Managers are created
	if err := core_managers.EnsureCAs(ctx, r.CaManagers, meshResource, meshResource.Meta.GetName()); err != nil {
		log.Error(err, "unable to ensure that mesh CAs are created")
		return kube_ctrl.Result{}, err
	}

	return kube_ctrl.Result{}, nil
}

func (r *MeshReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&mesh_k8s.Mesh{}).
		Complete(r)
}
