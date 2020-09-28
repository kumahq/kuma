package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8scnicncfio "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/apis/k8s.cni.cncf.io"
	network_v1 "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/apis/k8s.cni.cncf.io/v1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	kube_client.Client
	Log logr.Logger

	SystemNamespace         string
	CNIEnabled              bool
	SkipDefaultMeshCreation bool
	ResourceManager         core_manager.ResourceManager
	DefaultMeshTemplate     mesh_proto.Mesh
}

// Reconcile is in charge of creating NetworkAttachmentDefinition if CNI enabled and namespace has label 'kuma.io/sidecar-injection: enabled'
func (r *NamespaceReconciler) Reconcile(req kube_ctrl.Request) (kube_ctrl.Result, error) {
	if !r.CNIEnabled {
		return kube_ctrl.Result{}, nil
	}
	log := r.Log.WithValues("namespace", req.Name)
	ctx := context.Background()

	// Fetch the Namespace instance
	ns := &kube_core.Namespace{}
	if err := r.Get(ctx, req.NamespacedName, ns); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Namespace")
		return kube_ctrl.Result{}, err
	}

	if label, ok := ns.Labels["kuma.io/sidecar-injection"]; ok && label == "enabled" {
		err := r.createOrUpdateNetworkAttachmentDefinition(req.Name)
		return kube_ctrl.Result{}, err
	}
	return kube_ctrl.Result{}, nil
}

func (r *NamespaceReconciler) createOrUpdateNetworkAttachmentDefinition(namespace string) error {
	nad := &network_v1.NetworkAttachmentDefinition{
		ObjectMeta: kube_meta.ObjectMeta{
			Namespace: namespace,
			Name:      metadata.KumaCNI,
		},
	}
	_, err := kube_controllerutil.CreateOrUpdate(context.Background(), r.Client, nad, func() error {
		return nil
	})

	return err
}

func (r *NamespaceReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	if err := kube_core.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", kube_core.SchemeGroupVersion)
	}
	if err := mesh_k8s.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", mesh_k8s.GroupVersion)
	}
	if err := k8scnicncfio.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", k8scnicncfio.GroupVersion)
	}
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.Namespace{}).
		Watches(&kube_source.Kind{Type: &mesh_k8s.Mesh{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: kube_handler.ToRequestsFunc(func(obj kube_handler.MapObject) []kube_reconcile.Request {
				return []kube_reconcile.Request{{
					NamespacedName: kube_types.NamespacedName{Name: obj.Meta.GetNamespace()},
				}}
			}),
		}).
		Complete(r)
}
