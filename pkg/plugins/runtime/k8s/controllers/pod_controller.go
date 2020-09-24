package controllers

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/resources/model"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"

	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_record "k8s.io/client-go/tools/record"

	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	metadata "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"

	util_k8s "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

const (
	// CreatedKumaDataplaneReason is added to an event when
	// a new Dataplane is successfully created.
	CreatedKumaDataplaneReason = "CreatedKumaDataplane"
	// UpdatedKumaDataplaneReason is added to an event when
	// an existing Dataplane is successfully updated.
	UpdatedKumaDataplaneReason = "UpdatedKumaDataplane"
	// FailedToGenerateKumaDataplaneReason is added to an event when
	// a Dataplane cannot be generated or is not valid.
	FailedToGenerateKumaDataplaneReason = "FailedToGenerateKumaDataplane"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	kube_client.Client
	kube_record.EventRecorder
	Scheme       *kube_runtime.Scheme
	Log          logr.Logger
	PodConverter PodConverter
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

	// skip a Pod if it doesn't have an IP address yet
	if pod.Status.PodIP == "" {
		return kube_ctrl.Result{}, nil
	}

	// for Pods marked with ingress annotation special type of Dataplane will be injected
	enabled, exist, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaIngressAnnotation)
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	if exist && enabled {
		services, err := r.findMatchingServices(pod)
		if err != nil {
			return kube_ctrl.Result{}, err
		}
		return kube_ctrl.Result{}, r.createOrUpdateIngress(pod, services)
	}

	// only Pods with injected Kuma need a Dataplane descriptor
	injected, exist, err := metadata.Annotations(pod.Annotations).GetBool(metadata.KumaSidecarInjectedAnnotation)
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	if !exist || !injected {
		return kube_ctrl.Result{}, nil
	}

	services, err := r.findMatchingServices(pod)
	if err != nil {
		return kube_ctrl.Result{}, err
	}

	others, err := r.findOtherDataplanes(pod)
	if err != nil {
		return kube_ctrl.Result{}, err
	}

	r.Log.WithValues("req", req).V(1).Info("other dataplanes", "others", others)

	if err := r.createOrUpdateDataplane(pod, services, others); err != nil {
		return kube_ctrl.Result{}, err
	}

	return kube_ctrl.Result{}, nil
}

func (r *PodReconciler) findMatchingServices(pod *kube_core.Pod) ([]*kube_core.Service, error) {
	ctx := context.Background()

	// List Services in the same Namespace
	allServices := &kube_core.ServiceList{}
	if err := r.List(ctx, allServices, kube_client.InNamespace(pod.Namespace)); err != nil {
		log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
		log.Error(err, "unable to list Services", "namespace", pod.Namespace)
		return nil, err
	}

	// only consider Services that match this Pod
	matchingServices := util_k8s.FindServices(allServices, util_k8s.AnySelector(), util_k8s.MatchServiceThatSelectsPod(pod))

	return matchingServices, nil
}

func (r *PodReconciler) findOtherDataplanes(pod *kube_core.Pod) ([]*mesh_k8s.Dataplane, error) {
	ctx := context.Background()

	// List all Dataplanes
	allDataplanes := &mesh_k8s.DataplaneList{}
	if err := r.List(ctx, allDataplanes); err != nil {
		log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
		log.Error(err, "unable to list Dataplanes")
		return nil, err
	}

	// only consider Dataplanes in the same Mesh as Pod
	mesh := MeshFor(pod)
	otherDataplanes := make([]*mesh_k8s.Dataplane, 0)
	for i := range allDataplanes.Items {
		dataplane := allDataplanes.Items[i]
		if dataplane.Mesh == mesh {
			otherDataplanes = append(otherDataplanes, &dataplane)
		}
	}

	return otherDataplanes, nil
}

func (r *PodReconciler) createOrUpdateDataplane(pod *kube_core.Pod, services []*kube_core.Service, others []*mesh_k8s.Dataplane) error {
	ctx := context.Background()

	dataplane := &mesh_k8s.Dataplane{
		ObjectMeta: kube_meta.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		},
	}
	operationResult, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, dataplane, func() error {
		if err := r.PodConverter.PodToDataplane(dataplane, pod, services, others); err != nil {
			return errors.Wrap(err, "unable to translate a Pod into a Dataplane")
		}
		if err := kube_controllerutil.SetControllerReference(pod, dataplane, r.Scheme); err != nil {
			return errors.Wrap(err, "unable to set Dataplane's controller reference to Pod")
		}
		return nil
	})
	if err != nil {
		log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
		log.Error(err, "unable to create/update Dataplane", "operationResult", operationResult)
		r.EventRecorder.Eventf(pod, kube_core.EventTypeWarning, FailedToGenerateKumaDataplaneReason, "Failed to generate Kuma Dataplane: %s", err.Error())
		return err
	}
	switch operationResult {
	case kube_controllerutil.OperationResultCreated:
		r.EventRecorder.Eventf(pod, kube_core.EventTypeNormal, CreatedKumaDataplaneReason, "Created Kuma Dataplane: %s", pod.Name)
	case kube_controllerutil.OperationResultUpdated:
		r.EventRecorder.Eventf(pod, kube_core.EventTypeNormal, UpdatedKumaDataplaneReason, "Updated Kuma Dataplane: %s", pod.Name)
	}
	return nil
}

func (r *PodReconciler) createOrUpdateIngress(pod *kube_core.Pod, services []*kube_core.Service) error {
	ctx := context.Background()

	ingress := &mesh_k8s.Dataplane{
		ObjectMeta: kube_meta.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		},
		Mesh: model.DefaultMesh,
	}
	operationResult, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, ingress, func() error {
		if err := r.PodConverter.PodToIngress(ingress, pod, services); err != nil {
			return errors.Wrap(err, "unable to translate a Pod into a Ingress")
		}
		if err := kube_controllerutil.SetControllerReference(pod, ingress, r.Scheme); err != nil {
			return errors.Wrap(err, "unable to set Ingress's controller reference to Pod")
		}
		return nil
	})
	if err != nil {
		log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
		log.Error(err, "unable to create/update Ingress", "operationResult", operationResult)
		r.EventRecorder.Eventf(pod, kube_core.EventTypeWarning, FailedToGenerateKumaDataplaneReason, "Failed to generate Kuma Ingress: %s", err.Error())
		return err
	}
	switch operationResult {
	case kube_controllerutil.OperationResultCreated:
		r.EventRecorder.Eventf(pod, kube_core.EventTypeNormal, CreatedKumaDataplaneReason, "Created Kuma Ingress: %s", pod.Name)
	case kube_controllerutil.OperationResultUpdated:
		r.EventRecorder.Eventf(pod, kube_core.EventTypeNormal, UpdatedKumaDataplaneReason, "Updated Kuma Ingress: %s", pod.Name)
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
		// on Service update reconcile affected Pods (all Pods in the same namespace)
		Watches(&kube_source.Kind{Type: &kube_core.Service{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &ServiceToPodsMapper{Client: mgr.GetClient(), Log: r.Log.WithName("service-to-pods-mapper")},
		}).
		// on Dataplane update reconcile other Dataplanes in the same Mesh (ineffective, but that's the cost of not having Service abstraction)
		Watches(&kube_source.Kind{Type: &mesh_k8s.Dataplane{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &DataplaneToSameMeshDataplanesMapper{Client: mgr.GetClient(), Log: r.Log.WithName("dataplane-to-dataplanes-mapper")},
		}).
		Complete(r)
}

type ServiceToPodsMapper struct {
	kube_client.Client
	Log logr.Logger
}

func (m *ServiceToPodsMapper) Map(obj kube_handler.MapObject) []kube_reconile.Request {
	// List Pods in the same namespace as a Service
	pods := &kube_core.PodList{}
	if err := m.Client.List(context.Background(), pods, kube_client.InNamespace(obj.Meta.GetNamespace())); err != nil {
		m.Log.WithValues("service", obj.Meta).Error(err, "failed to fetch Pods")
		return nil
	}

	var req []kube_reconile.Request
	for _, pod := range pods.Items {
		req = append(req, kube_reconile.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name},
		})
	}
	return req
}

type DataplaneToSameMeshDataplanesMapper struct {
	kube_client.Client
	Log logr.Logger
}

func (m *DataplaneToSameMeshDataplanesMapper) Map(obj kube_handler.MapObject) []kube_reconile.Request {
	cause, ok := obj.Object.(*mesh_k8s.Dataplane)
	if !ok {
		m.Log.WithValues("dataplane", obj.Meta).Error(errors.Errorf("wrong argument type: expected %T, got %T", cause, obj.Object), "wrong argument type")
		return nil
	}

	ctx := context.Background()

	// Fetch the Dataplane instance
	if err := m.Client.Get(ctx, kube_types.NamespacedName{Namespace: cause.Namespace, Name: cause.Name}, &mesh_k8s.Dataplane{}); err != nil {
		if kube_apierrs.IsNotFound(err) {
			// a Dataplane object might be deleted by a user.
			// in that case we need to trigger reconciliation of a parent Pod.
			ownerRef := kube_meta.GetControllerOf(cause)
			if ownerRef == nil || ownerRef.Kind != "Pod" {
				return nil
			}
			return []kube_reconile.Request{
				{NamespacedName: kube_types.NamespacedName{Namespace: cause.Namespace, Name: ownerRef.Name}},
			}
		}
		m.Log.WithValues("dataplane", cause).Error(err, "failed to fetch Dataplane")
		return nil
	}

	// List Dataplanes in the same Mesh as the original
	dataplanes := &mesh_k8s.DataplaneList{}
	if err := m.Client.List(ctx, dataplanes); err != nil {
		m.Log.WithValues("dataplane", obj.Meta).Error(err, "failed to fetch Dataplanes")
		return nil
	}

	var req []kube_reconile.Request
	for _, dataplane := range dataplanes.Items {
		// skip Dataplanes from other Meshes
		if dataplane.Mesh != cause.Mesh {
			continue
		}
		// skip itself
		if dataplane.Namespace == cause.Namespace && dataplane.Name == cause.Name {
			continue
		}
		ownerRef := kube_meta.GetControllerOf(&dataplane)
		if ownerRef == nil || ownerRef.Kind != "Pod" {
			continue
		}
		req = append(req, kube_reconile.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: dataplane.Namespace, Name: ownerRef.Name},
		})
	}
	return req
}
