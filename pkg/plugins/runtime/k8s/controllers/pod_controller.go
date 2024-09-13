package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_discovery "k8s.io/api/discovery/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_record "k8s.io/client-go/tools/record"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	util_k8s "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
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
	Scheme                       *kube_runtime.Scheme
	Log                          logr.Logger
	PodConverter                 PodConverter
	ResourceConverter            k8s_common.Converter
	SystemNamespace              string
	IgnoredServiceSelectorLabels []string
}

func (r *PodReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	log := r.Log.WithValues("pod", req.NamespacedName)
	log.V(1).Info("reconcile")

	// Fetch the Pod instance
	pod := &kube_core.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		if kube_apierrs.IsNotFound(err) {
			log.V(1).Info("pod not found. Skipping")
			return kube_ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Pod")
		return kube_ctrl.Result{}, err
	}

	// for Pods marked with ingress annotation special type of Dataplane will be injected
	enabled, exist, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaIngressAnnotation)
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	if exist && enabled {
		return kube_ctrl.Result{}, r.reconcileZoneIngress(ctx, pod, log)
	}

	// for Pods marked with egress annotation special type of Dataplane will be injected
	egressEnabled, egressExist, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaEgressAnnotation)
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	if egressExist && egressEnabled {
		return kube_ctrl.Result{}, r.reconcileZoneEgress(ctx, pod, log)
	}

	// If we are using a builtin gateway, we want to generate a builtin gateway
	// dataplane.
	if name, _ := metadata.Annotations(pod.Annotations).GetString(metadata.KumaGatewayAnnotation); name == metadata.AnnotationBuiltin {
		return kube_ctrl.Result{}, r.reconcileBuiltinGatewayDataplane(ctx, pod, log)
	}

	// only Pods with injected Kuma need a Dataplane descriptor
	injected, _, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaSidecarInjectedAnnotation)
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	if injected {
		return kube_ctrl.Result{}, r.reconcileDataplane(ctx, pod, log)
	}

	return kube_ctrl.Result{}, nil
}

func (r *PodReconciler) reconcileDataplane(ctx context.Context, pod *kube_core.Pod, log logr.Logger) error {
	ns := kube_core.Namespace{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: pod.Namespace}, &ns); err != nil {
		return errors.Wrap(err, "unable to get Namespace for Pod")
	}

	if ns.Status.Phase == kube_core.NamespaceTerminating {
		r.Log.V(1).Info("namespace is terminating. Ignoring reconciliation")
		return nil
	}

	dp := &mesh_k8s.Dataplane{
		ObjectMeta: kube_meta.ObjectMeta{Name: pod.Name, Namespace: pod.Namespace},
	}
	if pod.Status.Phase == kube_core.PodSucceeded {
		// Remove Dataplane object for Pods that are indefinitely in Succeeded phase, i.e. Jobs
		return r.deleteObjectIfExist(ctx, dp, "pod succeeded", log)
	}
	if pod.Status.PodIP == "" {
		return r.deleteObjectIfExist(ctx, dp, "pod IP is empty", log)
	}
	if pod.Status.Reason == "Evicted" {
		return r.deleteObjectIfExist(ctx, dp, "pod was evicted", log)
	}

	services, err := r.findMatchingServices(ctx, pod)
	if err != nil {
		return err
	}

	var others []*mesh_k8s.Dataplane
	// we don't need other Dataplane objects when outbounds are stored in ConfigMap
	if !r.PodConverter.KubeOutboundsAsVIPs {
		others, err = r.findOtherDataplanes(ctx, pod, &ns)
		if err != nil {
			return err
		}
	}

	if err := r.createOrUpdateDataplane(ctx, pod, &ns, services, others); err != nil {
		return err
	}
	return nil
}

func (r *PodReconciler) deleteObjectIfExist(ctx context.Context, object k8s_model.KubernetesObject, cause string, log logr.Logger) error {
	log = log.WithValues(
		"cause", cause,
		"kind", object.GetObjectKind(),
		"name", object.GetName(),
		"namespace", object.GetNamespace(),
	)
	if err := r.Client.Delete(ctx, object); err != nil {
		if kube_apierrs.IsNotFound(err) {
			log.V(1).Info("Object is not found, nothing to delete")
			return nil
		}
		return errors.Wrapf(err, "could not delete %v %s/%s", object.GetObjectKind(), object.GetName(), object.GetNamespace())
	}
	log.Info("Object deleted")
	return nil
}

func (r *PodReconciler) reconcileBuiltinGatewayDataplane(ctx context.Context, pod *kube_core.Pod, log logr.Logger) error {
	if pod.Status.PodIP == "" {
		dp := &mesh_k8s.Dataplane{
			ObjectMeta: kube_meta.ObjectMeta{Name: pod.Name, Namespace: pod.Namespace},
		}
		return r.deleteObjectIfExist(ctx, dp, "pod IP is empty", log)
	}

	ns := kube_core.Namespace{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: pod.Namespace}, &ns); err != nil {
		return errors.Wrap(err, "unable to get Namespace for Pod")
	}
	return r.createorUpdateBuiltinGatewayDataplane(ctx, pod, &ns)
}

func (r *PodReconciler) reconcileZoneIngress(ctx context.Context, pod *kube_core.Pod, log logr.Logger) error {
	if pod.Status.PodIP == "" {
		zi := &mesh_k8s.ZoneIngress{
			ObjectMeta: kube_meta.ObjectMeta{Name: pod.Name, Namespace: pod.Namespace},
		}
		return r.deleteObjectIfExist(ctx, zi, "pod IP is empty", log)
	}

	if pod.Namespace != r.SystemNamespace {
		return errors.Errorf("Ingress can only be deployed in system namespace %q", r.SystemNamespace)
	}
	services, err := r.findMatchingServices(ctx, pod)
	if err != nil {
		return err
	}
	err = r.createOrUpdateIngress(ctx, pod, services)
	if err != nil {
		return err
	}
	return nil
}

func (r *PodReconciler) reconcileZoneEgress(ctx context.Context, pod *kube_core.Pod, log logr.Logger) error {
	if pod.Status.PodIP == "" {
		zi := &mesh_k8s.ZoneEgress{
			ObjectMeta: kube_meta.ObjectMeta{Name: pod.Name, Namespace: pod.Namespace},
		}
		return r.deleteObjectIfExist(ctx, zi, "pod IP is empty", log)
	}

	if pod.Namespace != r.SystemNamespace {
		return errors.Errorf("Egress can only be deployed in system namespace %q", r.SystemNamespace)
	}
	services, err := r.findMatchingServices(ctx, pod)
	if err != nil {
		return err
	}
	err = r.createOrUpdateEgress(ctx, pod, services)
	if err != nil {
		return err
	}
	return nil
}

func (r *PodReconciler) findByEndpointSlices(ctx context.Context, svc *kube_core.Service, objUID kube_types.UID) (bool, error) {
	endpointSlices := &kube_discovery.EndpointSliceList{}
	if err := r.List(
		ctx,
		endpointSlices,
		kube_client.InNamespace(svc.Namespace),
		kube_client.MatchingLabels(map[string]string{
			kube_discovery.LabelServiceName: svc.Name,
		}),
	); err != nil {
		return false, errors.Wrap(err, "unable to list EndpointSlices for service")
	}
	for _, slice := range endpointSlices.Items {
		for _, endpoint := range slice.Endpoints {
			if endpoint.TargetRef != nil && endpoint.TargetRef.UID == objUID {
				return true, nil
			}
		}
	}
	return false, nil
}

func (r *PodReconciler) findMatchingServices(ctx context.Context, pod *kube_core.Pod) ([]*kube_core.Service, error) {
	// List Services in the same Namespace
	allServices := &kube_core.ServiceList{}
	if err := r.List(ctx, allServices, kube_client.InNamespace(pod.Namespace)); err != nil {
		log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
		log.Error(err, "unable to list Services", "namespace", pod.Namespace)
		return nil, err
	}

	servicesByName := map[string]*kube_core.Service{}
	// only consider Services that match this Pod
	for svcIndex := range allServices.Items {
		svc := &allServices.Items[svcIndex]
		if util_k8s.MatchService(svc, util_k8s.Ignored()) {
			continue
		}
		if len(svc.Spec.Selector) == 0 {
			if endpointMatched, err := r.findByEndpointSlices(ctx, svc, pod.UID); err != nil {
				return nil, errors.Wrap(err, "unable to match services by EndpointSlices")
			} else if endpointMatched {
				servicesByName[svc.Name] = svc
			}
		} else if util_k8s.MatchServiceThatSelectsPod(pod, r.IgnoredServiceSelectorLabels)(svc) {
			servicesByName[svc.Name] = svc
		}
	}

	// Sort by name so that inbound order is similar across zones regardless of the order services got created.
	var matchingServices []*kube_core.Service
	for _, name := range util_maps.SortedKeys(servicesByName) {
		matchingServices = append(matchingServices, servicesByName[name])
	}
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
	mesh := util_k8s.MeshOfByAnnotation(pod, ns)
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
	log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
	if err != nil {
		if errors.Is(err, util_k8s.UnchangedResourceError) {
			log.V(1).Info("resource hasn't changed, skip")
			return nil
		}
		if !errors.Is(err, context.Canceled) {
			log.Error(err, "unable to create/update Dataplane", "operationResult", operationResult)
			r.EventRecorder.Eventf(pod, kube_core.EventTypeWarning, FailedToGenerateKumaDataplaneReason, "Failed to generate Kuma Dataplane: %s", err.Error())
		}

		return err
	}
	switch operationResult {
	case kube_controllerutil.OperationResultCreated:
		log.Info("Dataplane created")
		r.EventRecorder.Eventf(pod, kube_core.EventTypeNormal, CreatedKumaDataplaneReason, "Created Kuma Dataplane: %s", pod.Name)
	case kube_controllerutil.OperationResultUpdated:
		log.Info("Dataplane updated")
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
	log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
	if err != nil {
		if errors.Is(err, util_k8s.UnchangedResourceError) {
			log.V(1).Info("resource hasn't changed, skip")
			return nil
		}
		log.Error(err, "unable to create/update Ingress", "operationResult", operationResult)
		r.EventRecorder.Eventf(pod, kube_core.EventTypeWarning, FailedToGenerateKumaDataplaneReason, "Failed to generate Kuma Ingress: %s", err.Error())
		return err
	}
	switch operationResult {
	case kube_controllerutil.OperationResultCreated:
		log.Info("ZoneIngress created")
		r.EventRecorder.Eventf(pod, kube_core.EventTypeNormal, CreatedKumaDataplaneReason, "Created Kuma Ingress: %s", pod.Name)
	case kube_controllerutil.OperationResultUpdated:
		log.Info("ZoneIngress updated")
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
		if err := r.PodConverter.PodToEgress(ctx, egress, pod, services); err != nil {
			return errors.Wrap(err, "unable to translate a Pod into a Egress")
		}
		if err := kube_controllerutil.SetControllerReference(pod, egress, r.Scheme); err != nil {
			return errors.Wrap(err, "unable to set Egress's controller reference to Pod")
		}
		return nil
	})
	log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
	if err != nil {
		if errors.Is(err, util_k8s.UnchangedResourceError) {
			log.V(1).Info("resource hasn't changed, skip")
			return nil
		}
		log.Error(err, "unable to create/update Egress", "operationResult", operationResult)
		r.EventRecorder.Eventf(pod, kube_core.EventTypeWarning, FailedToGenerateKumaDataplaneReason, "Failed to generate Kuma Egress: %s", err.Error())
		return err
	}
	switch operationResult {
	case kube_controllerutil.OperationResultCreated:
		log.Info("ZoneEgress created")
		r.EventRecorder.Eventf(pod, kube_core.EventTypeNormal, CreatedKumaDataplaneReason, "Created Kuma Egress: %s", pod.Name)
	case kube_controllerutil.OperationResultUpdated:
		log.Info("ZoneEgress updated")
		r.EventRecorder.Eventf(pod, kube_core.EventTypeNormal, UpdatedKumaDataplaneReason, "Updated Kuma Egress: %s", pod.Name)
	}
	return nil
}

func (r *PodReconciler) SetupWithManager(mgr kube_ctrl.Manager, maxConcurrentReconciles int) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		Named("kuma-pod-controller").
		WithOptions(controller.Options{MaxConcurrentReconciles: maxConcurrentReconciles}).
		For(&kube_core.Pod{}).
		// on Service update reconcile affected Pods (all Pods selected by this service)
		Watches(&kube_core.Service{}, kube_handler.EnqueueRequestsFromMapFunc(ServiceToPodsMapper(r.Log, mgr.GetClient()))).
		Watches(&kube_discovery.EndpointSlice{}, kube_handler.EnqueueRequestsFromMapFunc(EndpointSliceToPodsMapper(r.Log, mgr.GetClient()))).
		Complete(r)
}

func ServiceToPodsMapper(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("service-to-pods-mapper")
	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		svc := obj.(*kube_core.Service)
		l := l.WithValues(
			"Service",
			kube_types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name},
		)
		// List Pods in the same namespace as a Service
		pods := &kube_core.PodList{}
		if len(svc.Spec.Selector) == 0 {
			endpointSlices := &kube_discovery.EndpointSliceList{}
			if err := client.List(
				ctx,
				endpointSlices,
				kube_client.InNamespace(svc.Namespace),
				kube_client.MatchingLabels(map[string]string{
					kube_discovery.LabelServiceName: svc.Name,
				}),
			); err != nil {
				l.Error(err, "failed to fetch EndpointSlices")
				return nil
			}
			var req []kube_reconcile.Request
			for _, slice := range endpointSlices.Items {
				for _, endpoint := range slice.Endpoints {
					if endpoint.TargetRef != nil && endpoint.TargetRef.Kind == "Pod" && (endpoint.TargetRef.APIVersion == kube_core.SchemeGroupVersion.String() ||
						endpoint.TargetRef.APIVersion == "") {
						req = append(req, kube_reconcile.Request{
							NamespacedName: kube_types.NamespacedName{Name: endpoint.TargetRef.Name, Namespace: endpoint.TargetRef.Namespace},
						})
					}
				}
			}
			return req
		}

		if err := client.List(ctx, pods, kube_client.InNamespace(obj.GetNamespace()), kube_client.MatchingLabels(svc.Spec.Selector)); err != nil {
			l.Error(err, "failed to fetch Pods matching selector")
			return nil
		}
		var req []kube_reconcile.Request
		for _, pod := range pods.Items {
			req = append(req, kube_reconcile.Request{
				NamespacedName: kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name},
			})
		}
		return req
	}
}

func EndpointSliceToPodsMapper(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("endpoint-slice-to-pods-mapper")
	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		slice := obj.(*kube_discovery.EndpointSlice)

		l := l.WithValues(
			"EndpointSlice",
			kube_types.NamespacedName{Namespace: slice.Namespace, Name: slice.Name},
		)

		svcName, ok := slice.Labels[kube_discovery.LabelServiceName]
		if !ok {
			return nil
		}
		var req []kube_reconcile.Request
		for _, endpoint := range slice.Endpoints {
			l := l.WithValues(
				"Endpoint",
				endpoint,
			)

			l = l.WithValues(
				"Service",
				kube_types.NamespacedName{Namespace: slice.Namespace, Name: svcName},
			)

			svc := &kube_core.Service{}
			if err := client.Get(ctx, kube_types.NamespacedName{Namespace: slice.Namespace, Name: svcName}, svc); err != nil {
				if kube_apierrs.IsNotFound(err) {
					continue
				}
				l.Error(err, "failed to get Service")
				// We intentionally continue here because we can't confirm we
				// can skip this EndpointSlice
			} else if len(svc.Spec.Selector) > 0 {
				// We don't consider EndpointSlices for this kind of Service
				continue
			}

			if endpoint.TargetRef != nil && endpoint.TargetRef.Kind == "Pod" && (endpoint.TargetRef.APIVersion == kube_core.SchemeGroupVersion.String() ||
				endpoint.TargetRef.APIVersion == "") {
				req = append(req, kube_reconcile.Request{
					NamespacedName: kube_types.NamespacedName{Namespace: endpoint.TargetRef.Namespace, Name: endpoint.TargetRef.Name},
				})
			}
		}
		return req
	}
}
