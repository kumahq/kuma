package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	k8scnicncfio "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/apis/k8s.cni.cncf.io"
	network_v1 "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/apis/k8s.cni.cncf.io/v1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	kube_client.Client
	Log logr.Logger

	CNIEnabled bool
}

// Reconcile is in charge of creating NetworkAttachmentDefinition if CNI enabled and namespace has label 'kuma.io/sidecar-injection: enabled'
func (r *NamespaceReconciler) Reconcile(req kube_ctrl.Request) (kube_ctrl.Result, error) {
	if !r.CNIEnabled {
		return kube_ctrl.Result{}, nil
	}
	log := r.Log.WithValues("namespace", req.Name)
	ctx := context.Background()

	ns := &kube_core.Namespace{}
	if err := r.Get(ctx, req.NamespacedName, ns); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, errors.Wrapf(err, "unable to fetch Namespace %s", req.NamespacedName.Name)
	}

	if ns.Status.Phase == kube_core.NamespaceTerminating {
		// we should not try to create or delete resources on namespace with Terminating state, it will result in errors
		log.V(1).Info("namespace is Terminating")
		return kube_ctrl.Result{}, nil
	}

	injected, _, err := metadata.Annotations(ns.Annotations).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrapf(err, "unable to check sidecar injection annotation on namespace %s", ns.Name)
	}
	if injected {
		log.Info("creating NetworkAttachmentDefinition for CNI support")
		err := r.createOrUpdateNetworkAttachmentDefinition(req.Name)
		return kube_ctrl.Result{}, err
	} else {
		// either a namespace that just had its kuma.io/sidecar-injection annotation removed or a namespace that never had this annotation
		err := r.deleteNetworkAttachmentDefinition(log, req.Name)
		return kube_ctrl.Result{}, err
	}
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

func (r *NamespaceReconciler) deleteNetworkAttachmentDefinition(log logr.Logger, namespace string) error {
	nad := &network_v1.NetworkAttachmentDefinition{}
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      metadata.KumaCNI,
	}
	err := r.Client.Get(context.Background(), key, nad)
	switch {
	case err == nil:
		log.Info("deleting NetworkAttachmentDefinition")
		return r.Client.Delete(context.Background(), nad)
	case kube_apierrs.IsNotFound(err): // it means that namespace never had Kuma injected
		return nil
	default:
		return errors.Wrap(err, "could not fetch NetworkAttachmentDefinition")
	}
}

func (r *NamespaceReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	if err := kube_core.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", kube_core.SchemeGroupVersion)
	}
	if err := k8scnicncfio.CNIAddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", k8scnicncfio.CNIGroupVersion)
	}
	if err := k8scnicncfio.CRDAddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", k8scnicncfio.CRDGroupVersion)
	}
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.Namespace{}, builder.WithPredicates(namespaceEvents)).
		Complete(r)
}

// we only want status event updates
var namespaceEvents = predicate.Funcs{
	CreateFunc: func(event event.CreateEvent) bool {
		return true
	},
	DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
		return false
	},
	UpdateFunc: func(updateEvent event.UpdateEvent) bool {
		return true
	},
	GenericFunc: func(genericEvent event.GenericEvent) bool {
		return false
	},
}
