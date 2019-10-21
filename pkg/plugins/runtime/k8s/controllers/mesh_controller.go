package controllers

import (
	"context"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	builtin_ca "github.com/Kong/kuma/pkg/core/ca/builtin"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	k8s_resources "github.com/Kong/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"

	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// MeshReconciler reconciles a Mesh object
type MeshReconciler struct {
	kube_client.Client
	Reader kube_client.Reader
	Log    logr.Logger

	Scheme           *kube_runtime.Scheme
	Converter        k8s_resources.Converter
	BuiltinCaManager builtin_ca.BuiltinCaManager
	SystemNamespace  string
	ResourceManager  manager.ResourceManager
}

func (r *MeshReconciler) Reconcile(req kube_ctrl.Request) (kube_ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("mesh", req.NamespacedName)

	// Fetch the Mesh instance
	mesh := &mesh_k8s.Mesh{}
	if err := r.Get(ctx, req.NamespacedName, mesh); err != nil {
		if kube_apierrs.IsNotFound(err) {
			err := r.ResourceManager.Delete(ctx, &mesh_core.MeshResource{}, store.DeleteByKey(req.Namespace, req.Name, req.Name))
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

	if !meshResource.HasBuiltinCA() {
		return kube_ctrl.Result{}, nil
	}

	if err := r.BuiltinCaManager.Ensure(ctx, mesh.Name); err != nil {
		log.Error(err, "unable to create Builtin CA")
		return kube_ctrl.Result{}, err
	}

	secretName := kube_types.NamespacedName{Namespace: r.SystemNamespace, Name: r.BuiltinCaManager.GetSecretName(mesh.Name)}

	log = log.WithValues("secret", secretName)

	// Fetch Secret instance
	secret := &kube_core.Secret{}
	if err := r.Reader.Get(ctx, secretName, secret); err != nil {
		if kube_apierrs.IsNotFound(err) {
			log.Error(err, "Builtin CA is not found")
			return kube_ctrl.Result{}, err
		}
		log.Error(err, "unable to fetch Secret")
		return kube_ctrl.Result{}, err
	}

	if err := kube_controllerutil.SetControllerReference(mesh, secret, r.Scheme); err != nil {
		log.Error(err, "unable to set Secret's controller reference to Mesh")
		return kube_ctrl.Result{}, err
	}

	if err := r.Client.Update(ctx, secret); err != nil {
		log.Error(err, "unable to save Secret after updating controller reference")
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
