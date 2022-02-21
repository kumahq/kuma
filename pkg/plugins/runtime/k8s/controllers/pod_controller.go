package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_record "k8s.io/client-go/tools/record"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/dns/vips"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
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

func (r *PodReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
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
		services, err := r.findMatchingServices(ctx, pod)
		if err != nil {
			return kube_ctrl.Result{}, err
		}
		err = r.createOrUpdateIngress(ctx, pod, services)
		if err != nil {
			return kube_ctrl.Result{}, err
		}

		// backwards compatibility
		// During the upgrade to Kuma 1.2.0 a new Ingress pod may make a request to the old-versioned PodController.
		// This will inevitably cause a generation of both Dataplane Ingress resource and Zone Ingress resource.
		// That's why here we're deleting Dataplane Ingress resource.
		if err = r.deleteOldTypeIngressIfExists(ctx, req, log); err != nil {
			return kube_ctrl.Result{}, err
		}
		return kube_ctrl.Result{}, nil
	}

	// for Pods marked with egress annotation special type of Dataplane will be injected
	egressEnabled, egressExist, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaEgressAnnotation)
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	if egressExist && egressEnabled {
		if pod.Namespace != r.SystemNamespace {
			return kube_ctrl.Result{}, errors.Errorf("Egress can only be deployed in system namespace %q", r.SystemNamespace)
		}
		services, err := r.findMatchingServices(ctx, pod)
		if err != nil {
			return kube_ctrl.Result{}, err
		}
		err = r.createOrUpdateEgress(ctx, pod, services)
		if err != nil {
			return kube_ctrl.Result{}, err
		}

		return kube_ctrl.Result{}, nil
	}

	ns := kube_core.Namespace{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: pod.Namespace}, &ns); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to get Namespace for Pod")
	}

	// If we are using a builtin gateway, we want to generate a builtin gateway
	// dataplane.
	if name, _ := metadata.Annotations(pod.Annotations).GetString(metadata.KumaGatewayAnnotation); name == metadata.AnnotationBuiltin {
		return kube_ctrl.Result{}, r.createorUpdateBuiltinGatewayDataplane(ctx, pod, &ns)
	}

	// only Pods with injected Kuma need a Dataplane descriptor
	injected, exist, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaSidecarInjectedAnnotation)
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	if !exist || !injected {
		return kube_ctrl.Result{}, nil
	}

	services, err := r.findMatchingServices(ctx, pod)
	if err != nil {
		return kube_ctrl.Result{}, err
	}

	others, err := r.findOtherDataplanes(ctx, pod, &ns)
	if err != nil {
		return kube_ctrl.Result{}, err
	}

	r.Log.WithValues("req", req).V(1).Info("other dataplanes", "others", others)

	if err := r.createOrUpdateDataplane(ctx, pod, &ns, services, others); err != nil {
		return kube_ctrl.Result{}, err
	}

	return kube_ctrl.Result{}, nil
}

func (r *PodReconciler) deleteOldTypeIngressIfExists(ctx context.Context, req kube_ctrl.Request, log logr.Logger) error {
	oldDpIngress := &mesh_k8s.Dataplane{}
	if err := r.Get(ctx, req.NamespacedName, oldDpIngress); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return nil
		}
		return err
	}

	log.Info("deleting dataplane-based Ingress, since new ZoneIngress resource was successfully created")
	if err := r.Delete(ctx, oldDpIngress); err != nil {
		log.Error(err, "failed to delete dataplane-based Ingress")
		return err
	}
	return nil
}

func (r *PodReconciler) findMatchingServices(ctx context.Context, pod *kube_core.Pod) ([]*kube_core.Service, error) {
	// List Services in the same Namespace
	allServices := &kube_core.ServiceList{}
	if err := r.List(ctx, allServices, kube_client.InNamespace(pod.Namespace)); err != nil {
		log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
		log.Error(err, "unable to list Services", "namespace", pod.Namespace)
		return nil, err
	}

	// only consider Services that match this Pod
	matchingServices := util_k8s.FindServices(allServices, util_k8s.Not(util_k8s.Ignored()), util_k8s.AnySelector(), util_k8s.MatchServiceThatSelectsPod(pod))

	return matchingServices, nil
}

func (r *PodReconciler) findOtherDataplanes(ctx context.Context, pod *kube_core.Pod, ns *kube_core.Namespace) ([]*mesh_k8s.Dataplane, error) {
	// List all Dataplanes
	allDataplanes := &mesh_k8s.DataplaneList{}
	if err := r.List(ctx, allDataplanes); err != nil {
		log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
		log.Error(err, "unable to list Dataplanes")
		return nil, err
	}

	// only consider Dataplanes in the same Mesh as Pod
	mesh := util_k8s.MeshOf(pod, ns)
	otherDataplanes := make([]*mesh_k8s.Dataplane, 0)
	for i := range allDataplanes.Items {
		dataplane := allDataplanes.Items[i]
		dp := core_mesh.NewDataplaneResource()
		if err := r.ResourceConverter.ToCoreResource(&dataplane, dp); err != nil {
			converterLog.Error(err, "failed to parse Dataplane", "dataplane", dataplane.Spec)
			continue // one invalid Dataplane definition should not break the entire mesh
		}
		if dataplane.Mesh == mesh {
			otherDataplanes = append(otherDataplanes, &dataplane)
		}
	}

	return otherDataplanes, nil
}

func (r *PodReconciler) createOrUpdateDataplane(
	ctx context.Context,
	pod *kube_core.Pod,
	ns *kube_core.Namespace,
	services []*kube_core.Service,
	others []*mesh_k8s.Dataplane,
) error {
	dataplane := &mesh_k8s.Dataplane{
		ObjectMeta: kube_meta.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		},
	}
	operationResult, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, dataplane, func() error {
		if err := r.PodConverter.PodToDataplane(ctx, dataplane, pod, ns, services, others); err != nil {
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

func (r *PodReconciler) createOrUpdateIngress(ctx context.Context, pod *kube_core.Pod, services []*kube_core.Service) error {
	ingress := &mesh_k8s.ZoneIngress{
		ObjectMeta: kube_meta.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		},
		Mesh: model.NoMesh,
	}
	operationResult, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, ingress, func() error {
		if err := r.PodConverter.PodToIngress(ctx, ingress, pod, services); err != nil {
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

func (r *PodReconciler) createOrUpdateEgress(ctx context.Context, pod *kube_core.Pod, services []*kube_core.Service) error {
	egress := &mesh_k8s.ZoneEgress{
		ObjectMeta: kube_meta.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		},
		Mesh: model.NoMesh,
	}
	operationResult, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, egress, func() error {
		if err := r.PodConverter.PodToEgress(egress, pod, services); err != nil {
			return errors.Wrap(err, "unable to translate a Pod into a Egress")
		}
		if err := kube_controllerutil.SetControllerReference(pod, egress, r.Scheme); err != nil {
			return errors.Wrap(err, "unable to set Egress's controller reference to Pod")
		}
		return nil
	})
	if err != nil {
		log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
		log.Error(err, "unable to create/update Egress", "operationResult", operationResult)
		r.EventRecorder.Eventf(pod, kube_core.EventTypeWarning, FailedToGenerateKumaDataplaneReason, "Failed to generate Kuma Egress: %s", err.Error())
		return err
	}
	switch operationResult {
	case kube_controllerutil.OperationResultCreated:
		r.EventRecorder.Eventf(pod, kube_core.EventTypeNormal, CreatedKumaDataplaneReason, "Created Kuma Egress: %s", pod.Name)
	case kube_controllerutil.OperationResultUpdated:
		r.EventRecorder.Eventf(pod, kube_core.EventTypeNormal, UpdatedKumaDataplaneReason, "Updated Kuma Egress: %s", pod.Name)
	}
	return nil
}

func (r *PodReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.Pod{}).
		// on Service update reconcile affected Pods (all Pods in the same namespace)
		Watches(&kube_source.Kind{Type: &kube_core.Service{}}, kube_handler.EnqueueRequestsFromMapFunc(ServiceToPodsMapper(r.Log, mgr.GetClient()))).
		// on ExternalService update reconcile affected Pods (all Pods in the same mesh)
		Watches(&kube_source.Kind{Type: &mesh_k8s.ExternalService{}}, kube_handler.EnqueueRequestsFromMapFunc(ExternalServiceToPodsMapper(r.Log, mgr.GetClient()))).
		Watches(&kube_source.Kind{Type: &kube_core.ConfigMap{}}, kube_handler.EnqueueRequestsFromMapFunc(ConfigMapToPodsMapper(r.Log, r.SystemNamespace, mgr.GetClient()))).
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

func ServiceToPodsMapper(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("service-to-pods-mapper")
	return func(obj kube_client.Object) []kube_reconile.Request {
		// List Pods in the same namespace as a Service
		pods := &kube_core.PodList{}
		if err := client.List(context.Background(), pods, kube_client.InNamespace(obj.GetNamespace())); err != nil {
			l.WithValues("service", obj.GetName()).Error(err, "failed to fetch Pods")
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
}

func ExternalServiceToPodsMapper(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("external-service-to-pods-mapper")
	return func(obj kube_client.Object) []kube_reconile.Request {
		cause, ok := obj.(*mesh_k8s.ExternalService)
		if !ok {
			l.WithValues("externalService", obj.GetName()).Error(errors.Errorf("wrong argument type: expected %T, got %T", cause, obj), "wrong argument type")
			return nil
		}

		// List Dataplanes in the same Mesh as the original
		dataplanes := &mesh_k8s.DataplaneList{}
		if err := client.List(context.Background(), dataplanes); err != nil {
			l.WithValues("dataplane", obj.GetName()).Error(err, "failed to fetch Dataplanes")
			return nil
		}

		var req []kube_reconile.Request
		for _, dataplane := range dataplanes.Items {
			// skip Dataplanes from other Meshes
			if dataplane.Mesh != cause.Mesh {
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
}

func ConfigMapToPodsMapper(l logr.Logger, ns string, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("configmap-to-pods-mapper")
	return func(obj kube_client.Object) []kube_reconile.Request {
		if obj.GetNamespace() != ns {
			return nil
		}
		mesh, ok := vips.MeshFromConfigKey(obj.GetName())
		if !ok {
			return nil
		}

		// List Dataplanes in the same Mesh as the original
		dataplanes := &mesh_k8s.DataplaneList{}
		if err := client.List(context.Background(), dataplanes); err != nil {
			l.WithValues("dataplane", obj.GetName()).Error(err, "failed to fetch Dataplanes")
			return nil
		}

		var req []kube_reconile.Request
		for _, dataplane := range dataplanes.Items {
			// skip Dataplanes from other Meshes
			if dataplane.Mesh != mesh {
				continue
			}
			// skip itself
			if dataplane.Namespace == obj.GetNamespace() && dataplane.Name == obj.GetName() {
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
}
