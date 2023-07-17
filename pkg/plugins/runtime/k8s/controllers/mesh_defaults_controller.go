package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_ctrl "sigs.k8s.io/controller-runtime"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	defaults_mesh "github.com/kumahq/kuma/pkg/defaults/mesh"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	common_k8s "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

// MeshDefaultsReconciler creates default resources for created Mesh
type MeshDefaultsReconciler struct {
	ResourceManager manager.ResourceManager
	Log             logr.Logger
}

func (r *MeshDefaultsReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	ctx = kuma_log.NewContext(ctx, kuma_log.FromContextOrDefault(ctx, r.Log))
	mesh := core_mesh.NewMeshResource()
	if err := r.ResourceManager.Get(ctx, mesh, store.GetByKey(req.Name, core_model.NoMesh)); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, errors.Wrap(err, "could not get default mesh resources")
	}

	// Before creating default policies for the mesh we want to ensure that this mesh wasn't processed before.
	if processed := mesh.GetMeta().(*k8s.KubernetesMetaAdapter).GetAnnotations()[common_k8s.K8sMeshDefaultsGenerated]; processed == "true" {
		return kube_ctrl.Result{}, nil
	}

	r.Log.Info("ensuring that default mesh resources exist", "mesh", req.Name)
	if err := defaults_mesh.EnsureDefaultMeshResources(ctx, r.ResourceManager, req.Name, mesh.Spec.GetSkipCreatingInitialPolicies()); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "could not create default mesh resources")
	}

	if mesh.GetMeta().(*k8s.KubernetesMetaAdapter).GetAnnotations() == nil {
		mesh.GetMeta().(*k8s.KubernetesMetaAdapter).Annotations = map[string]string{}
	}
	mesh.GetMeta().(*k8s.KubernetesMetaAdapter).GetAnnotations()[common_k8s.K8sMeshDefaultsGenerated] = "true"

	r.Log.Info("marking mesh that default resources were generated", "mesh", req.Name)

	if err := r.ResourceManager.Update(ctx, mesh); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "could not mark Mesh that default resources were generated")
	}
	return kube_ctrl.Result{}, nil
}

func (r *MeshDefaultsReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&mesh_k8s.Mesh{}).
		Complete(r)
}
