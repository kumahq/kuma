package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_managers "github.com/Kong/kuma/pkg/core/managers/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	kube_client.Client
	Log logr.Logger

	SystemNamespace     string
	ResourceManager     core_manager.ResourceManager
	DefaultMeshTemplate mesh_proto.Mesh
}

func (r *NamespaceReconciler) Reconcile(req kube_ctrl.Request) (kube_ctrl.Result, error) {
	if req.Name != r.SystemNamespace {
		return kube_ctrl.Result{}, nil
	}

	ctx := context.Background()
	log := r.Log.WithValues("namespace", req.Name)

	// Fetch the Namespace instance
	ns := &kube_core.Namespace{}
	if err := r.Get(ctx, req.NamespacedName, ns); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Namespace")
		return kube_ctrl.Result{}, err
	}

	// Fetch default Mesh instance
	mesh := &mesh_k8s.Mesh{}
	name := kube_types.NamespacedName{Name: core_model.DefaultMesh}
	if err := r.Get(ctx, name, mesh); err != nil {
		if kube_apierrs.IsNotFound(err) {
			err := mesh_managers.CreateDefaultMesh(r.ResourceManager, r.DefaultMeshTemplate)
			if err != nil {
				log.Error(err, "unable to create default Mesh")
			}
			return kube_ctrl.Result{}, err
		}
		log.Error(err, "unable to fetch Mesh", "name", name)
		return kube_ctrl.Result{}, err
	}

	return kube_ctrl.Result{}, nil
}

func (r *NamespaceReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	if err := kube_core.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", kube_core.SchemeGroupVersion)
	}
	if err := mesh_k8s.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", mesh_k8s.GroupVersion)
	}
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.Namespace{}).
		// on Mesh update reconcile Namespace it belongs to (in case default Mesh gets deleted)
		Watches(&kube_source.Kind{Type: &mesh_k8s.Mesh{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: kube_handler.ToRequestsFunc(func(obj kube_handler.MapObject) []kube_reconcile.Request {
				return []kube_reconcile.Request{{
					NamespacedName: kube_types.NamespacedName{Name: obj.Meta.GetNamespace()},
				}}
			}),
		}).
		Complete(r)
}
