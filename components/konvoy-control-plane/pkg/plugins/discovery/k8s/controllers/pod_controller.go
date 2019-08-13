package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	injector_metadata "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoy-injector/pkg/injector/metadata"
	mesh_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/api/v1alpha1"

	util_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/discovery/k8s/util"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *kube_runtime.Scheme
	Log    logr.Logger
}

func (r *PodReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pod", req.NamespacedName)

	// Fetch the Pod instance
	pod := &kube_core.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Pod")
		return ctrl.Result{}, err
	}

	// only Pods with injected Konvoy need a Dataplane descriptor
	if !injector_metadata.HasKonvoySidecar(pod) {
		return ctrl.Result{}, nil
	}

	// List Services in the same Namespace
	allServices := &kube_core.ServiceList{}
	if err := r.List(ctx, allServices, client.InNamespace(pod.Namespace)); err != nil {
		log.Error(err, "unable to list Services", "namespace", pod.Namespace)
		return ctrl.Result{}, err
	}

	// only consider Services that match this Pod
	matchingServices := util_k8s.FindServices(allServices, util_k8s.MatchServiceThatSelectsPod(pod))

	// create or update Dataplane definition
	dataplane := &mesh_k8s.Dataplane{
		ObjectMeta: kube_meta.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		},
	}
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, dataplane, func() error {
		if err := PodToDataplane(pod, matchingServices, dataplane); err != nil {
			return errors.Wrap(err, "unable to convert Pod to Dataplane")
		}
		if err := controllerutil.SetControllerReference(pod, dataplane, r.Scheme); err != nil {
			return errors.Wrap(err, "unable to set Dataplane's controller reference to Pod")
		}
		return nil
	})
	if err != nil {
		log.Error(err, "unable to create/update Dataplane", "op", op)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	for _, addToScheme := range []func(*kube_runtime.Scheme) error{kube_core.AddToScheme, mesh_k8s.AddToScheme} {
		if err := addToScheme(mgr.GetScheme()); err != nil {
			return err
		}
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.Pod{}).
		Complete(r)
}
