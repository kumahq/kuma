package controllers

import (
	"context"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/dns/vips"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"

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
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"

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
	Scheme            *kube_runtime.Scheme
	Log               logr.Logger
	PodConverter      PodConverter
	Persistence       *vips.Persistence
	ResourceConverter k8s_common.Converter
	SystemNamespace   string
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

	// skip a Pod if is complete/terminated (most probably a completed job)
	if r.isPodComplete(pod) {
		return kube_ctrl.Result{}, nil
	}

	// for Pods marked with ingress annotation special type of Dataplane will be injected
	enabled, exist, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaIngressAnnotation)
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	if exist && enabled {
		if pod.Namespace != r.SystemNamespace {
			return kube_ctrl.Result{}, errors.Errorf("Ingress can only be deployed in system namespace %q", r.SystemNamespace)
		}
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

	externalServices, err := r.findExternalServices(pod)
	if err != nil {
		return kube_ctrl.Result{}, err
	}

	others, err := r.findOtherDataplanes(pod)
	if err != nil {
		return kube_ctrl.Result{}, err
	}

	r.Log.WithValues("req", req).V(1).Info("other dataplanes", "others", others)

	vips, err := r.Persistence.GetByMesh(MeshFor(pod))
	if err != nil {
		return kube_ctrl.Result{}, err
	}

	if err := r.createOrUpdateDataplane(pod, services, externalServices, others, vips); err != nil {
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

func (r *PodReconciler) findExternalServices(pod *kube_core.Pod) ([]*mesh_k8s.ExternalService, error) {
	ctx := context.Background()

	// List all ExternalServices
	allExternalServices := &mesh_k8s.ExternalServiceList{}
	if err := r.List(ctx, allExternalServices); err != nil {
		log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
		log.Error(err, "unable to list ExternalServices")
		return nil, err
	}

	mesh := MeshFor(pod)
	meshedExternalServices := []*mesh_k8s.ExternalService{}
	for i := range allExternalServices.Items {
		es := allExternalServices.Items[i]
		if es.Mesh == mesh {
			meshedExternalServices = append(meshedExternalServices, &es)
		}
	}
	return meshedExternalServices, nil
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
		dp := core_mesh.NewDataplaneResource()
		if err := r.ResourceConverter.ToCoreResource(&dataplane, dp); err != nil {
			converterLog.Error(err, "failed to parse Dataplane", "dataplane", dataplane.Spec)
			continue // one invalid Dataplane definition should not break the entire mesh
		}
		if dataplane.Mesh == mesh || dp.Spec.IsIngress() {
			otherDataplanes = append(otherDataplanes, &dataplane)
		}
	}

	return otherDataplanes, nil
}

func (r *PodReconciler) createOrUpdateDataplane(
	pod *kube_core.Pod,
	services []*kube_core.Service,
	externalServices []*mesh_k8s.ExternalService,
	others []*mesh_k8s.Dataplane,
	vips vips.List,
) error {
	ctx := context.Background()

	dataplane := &mesh_k8s.Dataplane{
		ObjectMeta: kube_meta.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		},
	}
	operationResult, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, dataplane, func() error {
		if err := r.PodConverter.PodToDataplane(dataplane, pod, services, externalServices, others, vips); err != nil {
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
		Watches(&kube_source.Kind{Type: &kube_core.ConfigMap{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &ConfigMapToPodsMapper{Client: mgr.GetClient(), Log: r.Log.WithName("configmap-to-pods-mapper"), SystemNamespace: r.SystemNamespace},
		}).
		Complete(r)
}

func (r *PodReconciler) isPodComplete(pod *kube_core.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		// the sidecar amy or may not be terminated yet
		if cs.Name == util_k8s.KumaSidecarContainerName {
			continue
		}
		if cs.State.Terminated == nil {
			// at least one container not terminated, therefore pod is still active
			return false
		}
	}
	return true
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

type ConfigMapToPodsMapper struct {
	kube_client.Client
	Log             logr.Logger
	SystemNamespace string
}

func (m *ConfigMapToPodsMapper) Map(obj kube_handler.MapObject) []kube_reconile.Request {
	if obj.Meta.GetNamespace() != m.SystemNamespace {
		return nil
	}
	mesh, ok := vips.MeshFromConfigKey(obj.Meta.GetName())
	if !ok {
		return nil
	}

	// List Dataplanes in the same Mesh as the original
	dataplanes := &mesh_k8s.DataplaneList{}
	if err := m.Client.List(context.Background(), dataplanes); err != nil {
		m.Log.WithValues("dataplane", obj.Meta).Error(err, "failed to fetch Dataplanes")
		return nil
	}

	var req []kube_reconile.Request
	for _, dataplane := range dataplanes.Items {
		// skip Dataplanes from other Meshes
		if dataplane.Mesh != mesh {
			continue
		}
		// skip itself
		if dataplane.Namespace == obj.Meta.GetNamespace() && dataplane.Name == obj.Meta.GetName() {
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
