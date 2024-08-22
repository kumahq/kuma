package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_ctrl "sigs.k8s.io/controller-runtime"

	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_managers "github.com/kumahq/kuma/pkg/core/managers/apis/mesh"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	defaults_mesh "github.com/kumahq/kuma/pkg/defaults/mesh"
	common_k8s "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

// MeshReconciler creates default resources for created Mesh and ensures that CA was created
type MeshReconciler struct {
	ResourceManager            manager.ResourceManager
	Log                        logr.Logger
	Extensions                 context.Context
	CreateMeshRoutingResources bool
	K8sStore                   bool
	CaManagers                 core_ca.Managers
	SystemNamespace            string
}

func (r *MeshReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	mesh := core_mesh.NewMeshResource()
	if err := r.ResourceManager.Get(ctx, mesh, store.GetByKey(req.Name, core_model.NoMesh)); err != nil {
		if store.IsResourceNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, errors.Wrap(err, "could not get default mesh resources")
	}

	r.Log.V(1).Info("ensuring CAs for mesh exist")
	if err := core_managers.EnsureCAs(ctx, r.CaManagers, mesh, mesh.Meta.GetName()); err != nil {
		r.Log.Error(err, "unable to ensure that mesh CAs are created")
		return kube_ctrl.Result{}, err
	}

	if err := r.ensureDefaultResources(ctx, mesh); err != nil {
		return kube_ctrl.Result{}, err
	}

	return kube_ctrl.Result{}, nil
}

func (r *MeshReconciler) ensureDefaultResources(ctx context.Context, mesh *core_mesh.MeshResource) error {
	// Before creating default policies for the mesh we want to ensure that this mesh wasn't processed before.
	if processed := mesh.GetMeta().(*k8s.KubernetesMetaAdapter).GetAnnotations()[common_k8s.K8sMeshDefaultsGenerated]; processed == "true" {
		return nil
	}

	r.Log.Info("ensuring that default mesh resources exist", "mesh", mesh.GetMeta().GetName())
	if err := defaults_mesh.EnsureDefaultMeshResources(
		ctx,
		r.ResourceManager,
		mesh,
		mesh.Spec.GetSkipCreatingInitialPolicies(),
		r.Extensions,
		r.CreateMeshRoutingResources,
		r.K8sStore,
		r.SystemNamespace,
	); err != nil {
		return errors.Wrap(err, "could not create default mesh resources")
	}

	if mesh.GetMeta().(*k8s.KubernetesMetaAdapter).GetAnnotations() == nil {
		mesh.GetMeta().(*k8s.KubernetesMetaAdapter).Annotations = map[string]string{}
	}
	mesh.GetMeta().(*k8s.KubernetesMetaAdapter).GetAnnotations()[common_k8s.K8sMeshDefaultsGenerated] = "true"

	r.Log.Info("marking mesh that default resources were generated", "mesh", mesh.GetMeta().GetName())

	if err := r.ResourceManager.Update(ctx, mesh); err != nil {
		return errors.Wrap(err, "could not mark Mesh that default resources were generated")
	}
	return nil
}

func (r *MeshReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		Named("kuma-mesh-controller").
		For(&mesh_k8s.Mesh{}).
		Complete(r)
}
