package controllers

import (
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	defaults_mesh "github.com/kumahq/kuma/pkg/defaults/mesh"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

// MeshDefaultsReconciler creates default resources for a created Mesh
type MeshDefaultsReconciler struct {
	Log             logr.Logger
	ResourceManager manager.ResourceManager
}

func (r *MeshDefaultsReconciler) Reconcile(req kube_ctrl.Request) (kube_ctrl.Result, error) {
	log := r.Log.WithValues("mesh", req.NamespacedName)

	// we skip default name Mesh because pkg/defaults/mesh.go uses ResourceManager that executes CreateDefaultMeshResources already
	if req.Name == model.DefaultMesh {
		return kube_ctrl.Result{}, nil
	}

	if err := defaults_mesh.CreateDefaultMeshResources(r.ResourceManager, req.Name); err != nil {
		log.Error(err, "could not create default mesh resources")
	}
	return kube_ctrl.Result{}, nil
}

func (r *MeshDefaultsReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	if err := kube_core.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", kube_core.SchemeGroupVersion)
	}
	if err := mesh_k8s.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", mesh_k8s.GroupVersion)
	}
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&mesh_k8s.Mesh{}, builder.WithPredicates(onlyCreate)).
		Complete(r)
}

var onlyCreate = predicate.Funcs{
	CreateFunc: func(event event.CreateEvent) bool {
		return true
	},
	DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
		return false
	},
	UpdateFunc: func(updateEvent event.UpdateEvent) bool {
		return false
	},
	GenericFunc: func(genericEvent event.GenericEvent) bool {
		return false
	},
}
