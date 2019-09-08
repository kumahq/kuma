package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	injector_metadata "github.com/Kong/konvoy/components/konvoy-control-plane/app/kuma-injector/pkg/injector/metadata"
	mesh_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/api/v1alpha1"

	util_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/discovery/k8s/util"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	kube_client.Client
	Scheme *kube_runtime.Scheme
	Log    logr.Logger
}

func (r *PodReconciler) Reconcile(req kube_ctrl.Request) (kube_ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pod", req.NamespacedName)

	// Fetch the Pod instance
	pod := &kube_core.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Pod")
		return kube_ctrl.Result{}, err
	}

	// only Pods with injected Konvoy need a Dataplane descriptor
	if !injector_metadata.HasKumaSidecar(pod) {
		return kube_ctrl.Result{}, nil
	}

	// skip a Pod if it doesn't have an IP address yet
	if pod.Status.PodIP == "" {
		return kube_ctrl.Result{}, nil
	}

	services, err := r.matchingServices(pod)
	if err != nil {
		return kube_ctrl.Result{}, err
	}

	if err := r.createOrUpdateDataplane(pod, services); err != nil {
		return kube_ctrl.Result{}, err
	}

	return kube_ctrl.Result{}, nil
}

func (r *PodReconciler) matchingServices(pod *kube_core.Pod) ([]*kube_core.Service, error) {
	ctx := context.Background()

	// List Services in the same Namespace
	allServices := &kube_core.ServiceList{}
	if err := r.List(ctx, allServices, kube_client.InNamespace(pod.Namespace)); err != nil {
		log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
		log.Error(err, "unable to list Services", "namespace", pod.Namespace)
		return nil, err
	}

	// only consider Services that match this Pod
	matchingServices := util_k8s.FindServices(allServices, util_k8s.MatchServiceThatSelectsPod(pod))

	return matchingServices, nil
}

func (r *PodReconciler) createOrUpdateDataplane(pod *kube_core.Pod, services []*kube_core.Service) error {
	ctx := context.Background()

	dataplane := &mesh_k8s.Dataplane{
		ObjectMeta: kube_meta.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		},
	}
	operationResult, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, dataplane, func() error {
		if err := PodToDataplane(pod, services, dataplane); err != nil {
			return errors.Wrap(err, "unable to convert Pod to Dataplane")
		}
		if err := kube_controllerutil.SetControllerReference(pod, dataplane, r.Scheme); err != nil {
			return errors.Wrap(err, "unable to set Dataplane's controller reference to Pod")
		}
		return nil
	})
	if err != nil {
		log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
		log.Error(err, "unable to create/update Dataplane", "operationResult", operationResult)
		return err
	}
	return nil
}

func (r *PodReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	for _, addToScheme := range []func(*kube_runtime.Scheme) error{kube_core.AddToScheme, mesh_k8s.AddToScheme} {
		if err := addToScheme(mgr.GetScheme()); err != nil {
			return err
		}
	}
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.Pod{}).
		Complete(r)
}
