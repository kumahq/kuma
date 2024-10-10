package controllers

import (
	"context"
	"fmt"
	"maps"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_apps "k8s.io/api/apps/v1"
	kube_core "k8s.io/api/core/v1"
	kube_discovery "k8s.io/api/discovery/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_record "k8s.io/client-go/tools/record"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshservice_k8s "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/k8s/v1alpha1"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	"github.com/kumahq/kuma/pkg/util/k8s"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

const (
	// CreatedMeshServiceReason is added to an event when
	// a new MeshService is successfully created.
	CreatedMeshServiceReason = "CreatedMeshService"
	// UpdatedMeshServiceReason is added to an event when
	// an existing MeshService is successfully updated.
	UpdatedMeshServiceReason = "UpdatedMeshService"
	// FailedToGenerateMeshServiceReason is added to an event when
	// a MeshService cannot be generated.
	FailedToGenerateMeshServiceReason = "FailedToGenerateMeshService"
	IgnoredUnsupportedPortReason      = "IgnoredUnsupportedPort"
)

// MeshServiceReconciler reconciles a MeshService object
type MeshServiceReconciler struct {
	kube_client.Client
	kube_record.EventRecorder
	Log               logr.Logger
	Scheme            *kube_runtime.Scheme
	ResourceConverter k8s_common.Converter
}

func (r *MeshServiceReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	log := r.Log.WithValues("service", req.NamespacedName)

	namespace := &kube_core.Namespace{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: req.Namespace}, namespace); err != nil {
		if kube_apierrs.IsNotFound(err) {
			// MeshService will be deleted automatically.
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to get Namespace for Service")
	}
	injectedLabel, _, err := metadata.Annotations(namespace.Labels).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrapf(err, "unable to check sidecar injection label on namespace %s", namespace.Name)
	}

	svc := &kube_core.Service{}
	if err := r.Get(ctx, req.NamespacedName, svc); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, errors.Wrapf(err, "unable to fetch Service %s", req.NamespacedName.Name)
	}

	_, ok := svc.GetLabels()[mesh_proto.MeshTag]
	if !ok && !injectedLabel {
		log.V(1).Info("service is not considered to be service in a mesh")
		if err := r.deleteIfExist(ctx, req.NamespacedName); err != nil {
			return kube_ctrl.Result{}, err
		}
		return kube_ctrl.Result{}, nil
	}

	meshName := util.MeshOfByLabelOrAnnotation(log, svc, namespace)

	k8sMesh := v1alpha1.Mesh{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: meshName}, &k8sMesh); err != nil {
		if kube_apierrs.IsNotFound(err) {
			log.V(1).Info("mesh not found")
			if err := r.deleteIfExist(ctx, req.NamespacedName); err != nil {
				return kube_ctrl.Result{}, err
			}
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, err
	}

	mesh := core_mesh.NewMeshResource()
	if err := r.ResourceConverter.ToCoreResource(&k8sMesh, mesh); err != nil {
		return kube_ctrl.Result{}, err
	}

	if mesh.Spec.MeshServicesMode() == mesh_proto.Mesh_MeshServices_Disabled {
		log.V(1).Info("MeshServices not enabled on Mesh, deleting existing")
		if err := r.deleteIfExist(ctx, req.NamespacedName); err != nil {
			return kube_ctrl.Result{}, err
		}
		return kube_ctrl.Result{}, nil
	}

	if len(svc.GetAnnotations()) > 0 {
		if _, ok := svc.GetAnnotations()[metadata.KumaGatewayAnnotation]; ok {
			log.V(1).Info("service is for gateway. Ignoring.")
			return kube_ctrl.Result{}, nil
		}
	}

	if svc.Spec.ClusterIP == "" {
		log.V(1).Info("service has no cluster IP. Ignoring.")
		return kube_ctrl.Result{}, nil
	}

	if svc.Spec.ClusterIP != kube_core.ClusterIPNone {
		name := kube_client.ObjectKeyFromObject(svc)
		op, err := r.manageMeshService(
			ctx,
			svc,
			meshName,
			r.setFromClusterIPSvc,
			name,
		)
		if err != nil {
			return kube_ctrl.Result{}, err
		}
		switch op {
		case kube_controllerutil.OperationResultCreated:
			r.EventRecorder.Eventf(svc, kube_core.EventTypeNormal, CreatedMeshServiceReason, "Created Kuma MeshService: %s", name.Name)
		case kube_controllerutil.OperationResultUpdated:
			r.EventRecorder.Eventf(svc, kube_core.EventTypeNormal, UpdatedMeshServiceReason, "Updated Kuma MeshService: %s", name.Name)
		}

		return kube_ctrl.Result{}, nil
	}

	trackedPodEndpoints := map[kube_types.NamespacedName]struct{}{}
	meshServices := &meshservice_k8s.MeshServiceList{}
	if err := r.List(
		ctx,
		meshServices,
		kube_client.MatchingLabels(map[string]string{
			metadata.KumaServiceName: svc.Name,
		}),
	); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to list MeshServices for headless Service")
	}
	for _, svc := range meshServices.Items {
		if len(svc.GetOwnerReferences()) == 0 {
			continue
		}
		owner := svc.GetOwnerReferences()[0]
		if owner.Kind != "Pod" || owner.APIVersion != kube_core.SchemeGroupVersion.String() {
			continue
		}
		trackedPodEndpoints[kube_types.NamespacedName{Namespace: svc.Namespace, Name: owner.Name}] = struct{}{}
	}

	endpointSlices := &kube_discovery.EndpointSliceList{}
	if err := r.List(
		ctx,
		endpointSlices,
		kube_client.InNamespace(svc.Namespace),
		kube_client.MatchingLabels(map[string]string{
			kube_discovery.LabelServiceName: svc.Name,
		}),
	); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to list EndpointSlices for headless Service")
	}

	servicePodEndpoints := map[kube_types.NamespacedName]kube_discovery.Endpoint{}
	// We need to look at our EndpointSlice to see which Pods this headless
	// service points to
	var created, updated int
	for _, slice := range endpointSlices.Items {
		for _, endpoint := range slice.Endpoints {
			if endpoint.TargetRef == nil ||
				endpoint.TargetRef.Kind != "Pod" ||
				(endpoint.TargetRef.APIVersion != kube_core.SchemeGroupVersion.String() &&
					endpoint.TargetRef.APIVersion != "") {
				continue
			}
			servicePodEndpoints[kube_types.NamespacedName{Name: endpoint.TargetRef.Name, Namespace: endpoint.TargetRef.Namespace}] = endpoint
		}
	}

	// Delete trackedPodEndpoints - servicePodEndpoints
	for tracked := range trackedPodEndpoints {
		if _, ok := servicePodEndpoints[tracked]; ok {
			continue
		}
		delete(trackedPodEndpoints, tracked)
		ms := meshservice_k8s.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Namespace: tracked.Namespace,
				Name:      tracked.Name,
			},
		}
		if err := r.Delete(ctx, &ms); err != nil && !kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, errors.Wrap(err, "unable to delete MeshService tracking headless Service endpoint")
		}
	}

	for current, endpoint := range servicePodEndpoints {
		// Our name is unique depending on the service identity and pod name
		// Note that a Pod name can be the same length as a Service name
		// so we might need to truncate the MeshService name
		namePrefix := current.Name
		canonicalNameHasher := k8s.NewHasher()
		canonicalNameHasher.Write([]byte(svc.Name))
		canonicalNameHasher.Write([]byte(svc.Namespace))
		// name + `-` + 10 characters
		if len(current.Name)+k8s.MaxHashStringLength+1 > 63 {
			canonicalNameHasher.Write([]byte(current.Name))
			namePrefix = k8s.EnsureMaxLength(namePrefix, 63-k8s.MaxHashStringLength-1)
		}
		canonicalName := fmt.Sprintf("%s-%s", namePrefix, k8s.HashToString(canonicalNameHasher))
		op, err := r.manageMeshService(
			ctx,
			svc,
			meshName,
			r.setFromPodAndHeadlessSvc(endpoint),
			kube_types.NamespacedName{Namespace: current.Namespace, Name: canonicalName},
		)
		if err != nil {
			return kube_ctrl.Result{}, errors.Wrap(err, "unable to create/update MeshService for headless Service")
		}
		switch op {
		case kube_controllerutil.OperationResultCreated:
			created++
		case kube_controllerutil.OperationResultUpdated:
			updated++
		}
	}
	if created > 0 {
		r.EventRecorder.Eventf(svc, kube_core.EventTypeNormal, CreatedMeshServiceReason, "Created %d MeshServices", created)
	}
	if updated > 0 {
		r.EventRecorder.Eventf(svc, kube_core.EventTypeNormal, UpdatedMeshServiceReason, "Updated %d MeshServices", updated)
	}

	return kube_ctrl.Result{}, nil
}

func (r *MeshServiceReconciler) setFromClusterIPSvc(_ context.Context, ms *meshservice_k8s.MeshService, svc *kube_core.Service) error {
	if ms.ObjectMeta.GetGeneration() != 0 {
		if owners := ms.GetOwnerReferences(); len(owners) == 0 || owners[0].UID != svc.GetUID() {
			r.EventRecorder.Eventf(
				svc, kube_core.EventTypeWarning, FailedToGenerateMeshServiceReason, "MeshService already exists and isn't owned by Service",
			)
			return errors.Errorf("MeshService already exists and isn't owned by Service")
		}
	}
	ms.ObjectMeta.Labels[metadata.HeadlessService] = "false"
	dpTags := maps.Clone(svc.Spec.Selector)
	if dpTags == nil {
		dpTags = map[string]string{}
	}
	dpTags[mesh_proto.KubeNamespaceTag] = svc.GetNamespace()
	ms.Spec.Selector = meshservice_api.Selector{
		DataplaneTags: dpTags,
	}

	ms.Status.VIPs = []meshservice_api.VIP{
		{
			IP: svc.Spec.ClusterIP,
		},
	}

	if err := kube_controllerutil.SetOwnerReference(svc, ms, r.Scheme); err != nil {
		return errors.Wrap(err, "could not set owner reference")
	}
	return nil
}

func (r *MeshServiceReconciler) setFromPodAndHeadlessSvc(endpoint kube_discovery.Endpoint) func(context.Context, *meshservice_k8s.MeshService, *kube_core.Service) error {
	return func(ctx context.Context, ms *meshservice_k8s.MeshService, svc *kube_core.Service) error {
		if ms.ObjectMeta.GetGeneration() != 0 {
			if owners := ms.GetOwnerReferences(); len(owners) == 0 || owners[0].UID != endpoint.TargetRef.UID {
				r.EventRecorder.Eventf(
					svc, kube_core.EventTypeWarning, FailedToGenerateMeshServiceReason, "MeshService already exists and isn't owned by Pod",
				)
				return errors.Errorf("MeshService already exists and isn't owned by Pod")
			}
		}
		if ms.ObjectMeta.Labels == nil {
			ms.ObjectMeta.Labels = map[string]string{}
		}
		ms.ObjectMeta.Labels[metadata.HeadlessService] = "true"
		pod := kube_core.Pod{}
		if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: endpoint.TargetRef.Name, Namespace: endpoint.TargetRef.Namespace}, &pod); err != nil {
			if !kube_apierrs.IsNotFound(err) {
				return errors.Wrap(err, "couldn't lookup Pod for endpoint")
			}
		} else {
			if v, ok := pod.Labels[kube_apps.StatefulSetPodNameLabel]; ok {
				ms.ObjectMeta.Labels[kube_apps.StatefulSetPodNameLabel] = v
			}
			if v, ok := pod.Labels[kube_apps.PodIndexLabel]; ok {
				ms.ObjectMeta.Labels[kube_apps.PodIndexLabel] = v
			}
		}
		ms.Spec.Selector = meshservice_api.Selector{
			DataplaneRef: &meshservice_api.DataplaneRef{
				Name: fmt.Sprintf("%s.%s", endpoint.TargetRef.Name, endpoint.TargetRef.Namespace),
			},
		}
		var vips []meshservice_api.VIP
		for _, address := range endpoint.Addresses {
			vips = append(vips,
				meshservice_api.VIP{
					IP: address,
				})
		}
		ms.Status.VIPs = vips
		owner := kube_core.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      endpoint.TargetRef.Name,
				Namespace: endpoint.TargetRef.Namespace,
				UID:       endpoint.TargetRef.UID,
			},
		}
		if err := kube_controllerutil.SetOwnerReference(&owner, ms, r.Scheme); err != nil {
			return errors.Wrap(err, "could not set owner reference")
		}
		return nil
	}
}

func (r *MeshServiceReconciler) manageMeshService(
	ctx context.Context,
	svc *kube_core.Service,
	mesh string,
	setSpec func(context.Context, *meshservice_k8s.MeshService, *kube_core.Service) error,
	meshServiceName kube_types.NamespacedName,
) (kube_controllerutil.OperationResult, error) {
	ms := &meshservice_k8s.MeshService{
		ObjectMeta: v1.ObjectMeta{
			Name:      meshServiceName.Name,
			Namespace: meshServiceName.Namespace,
		},
	}

	var unsupportedPorts []string

	result, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, ms, func() error {
		ms.ObjectMeta.Labels = maps.Clone(svc.GetLabels())
		if ms.ObjectMeta.Labels == nil {
			ms.ObjectMeta.Labels = map[string]string{}
		}
		ms.ObjectMeta.Labels[mesh_proto.MeshTag] = mesh
		ms.ObjectMeta.Labels[metadata.KumaServiceName] = svc.GetName()
		ms.ObjectMeta.Labels[mesh_proto.ManagedByLabel] = "k8s-controller"
		ms.ObjectMeta.Labels[mesh_proto.EnvTag] = mesh_proto.KubernetesEnvironment

		if ms.Spec == nil {
			ms.Spec = &meshservice_api.MeshService{}
		}

		ms.Spec.Ports = []meshservice_api.Port{}
		for _, port := range svc.Spec.Ports {
			if port.Protocol != kube_core.ProtocolTCP {
				portName := port.Name
				if portName == "" {
					portName = strconv.Itoa(int(port.Port))
				}
				unsupportedPorts = append(unsupportedPorts, portName)
				continue
			}
			portName := port.Name
			if portName == "" {
				portName = strconv.Itoa(int(port.Port))
			}
			ms.Spec.Ports = append(ms.Spec.Ports, meshservice_api.Port{
				Name:        portName,
				Port:        uint32(port.Port),
				TargetPort:  port.TargetPort,
				AppProtocol: core_mesh.Protocol(pointer.DerefOr(port.AppProtocol, "tcp")),
			})
		}

		if ms.Status == nil {
			ms.Status = &meshservice_api.MeshServiceStatus{}
		}
		return setSpec(ctx, ms, svc)
	})

	if result != kube_controllerutil.OperationResultNone && len(unsupportedPorts) > 0 {
		ports := strings.Join(unsupportedPorts, ", ")
		r.EventRecorder.Eventf(svc, kube_core.EventTypeNormal, IgnoredUnsupportedPortReason, "Ignored unsupported ports: %s", ports)
	}

	return result, err
}

func (r *MeshServiceReconciler) deleteIfExist(ctx context.Context, key kube_types.NamespacedName) error {
	ms := &meshservice_k8s.MeshService{
		ObjectMeta: v1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
	}
	if err := r.Client.Delete(ctx, ms); err != nil && !kube_apierrs.IsNotFound(err) {
		return errors.Wrap(err, "could not delete MeshService")
	}
	return nil
}

func (r *MeshServiceReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		Named("kuma-mesh-service-controller").
		For(&kube_core.Service{}).
		Watches(&kube_core.Namespace{}, kube_handler.EnqueueRequestsFromMapFunc(NamespaceToServiceMapper(r.Log, mgr.GetClient())), builder.WithPredicates(predicate.LabelChangedPredicate{})).
		Watches(&v1alpha1.Mesh{}, kube_handler.EnqueueRequestsFromMapFunc(MeshToAllMeshServices(r.Log, mgr.GetClient())), builder.WithPredicates(CreateOrDeletePredicate{})).
		Watches(&kube_discovery.EndpointSlice{}, kube_handler.EnqueueRequestsFromMapFunc(EndpointSliceToServicesMapper(r.Log, mgr.GetClient()))).
		Complete(r)
}

func EndpointSliceToServicesMapper(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		slice := obj.(*kube_discovery.EndpointSlice)

		svcName, ok := slice.Labels[kube_discovery.LabelServiceName]
		if !ok {
			return nil
		}
		req := []kube_reconcile.Request{
			{NamespacedName: kube_types.NamespacedName{Namespace: slice.Namespace, Name: svcName}},
		}
		return req
	}
}

func NamespaceToServiceMapper(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("namespace-to-service-mapper")
	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		services := &kube_core.ServiceList{}
		if err := client.List(ctx, services, kube_client.InNamespace(obj.GetNamespace())); err != nil {
			l.WithValues("namespace", obj.GetName()).Error(err, "failed to fetch Services")
			return nil
		}
		var req []kube_reconcile.Request
		for _, svc := range services.Items {
			req = append(req, kube_reconcile.Request{
				NamespacedName: kube_types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name},
			})
		}
		return req
	}
}

func MeshToAllMeshServices(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("mesh-to-service-mapper")
	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		services := &kube_core.ServiceList{}
		if err := client.List(ctx, services); err != nil {
			l.WithValues("namespace", obj.GetName()).Error(err, "failed to fetch Services")
			return nil
		}
		var req []kube_reconcile.Request
		for _, svc := range services.Items {
			req = append(req, kube_reconcile.Request{
				NamespacedName: kube_types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name},
			})
		}
		return req
	}
}

type CreateOrDeletePredicate struct {
	predicate.Funcs
}

func (p CreateOrDeletePredicate) Create(e event.CreateEvent) bool {
	return true
}

func (p CreateOrDeletePredicate) Delete(e event.DeleteEvent) bool {
	return true
}
