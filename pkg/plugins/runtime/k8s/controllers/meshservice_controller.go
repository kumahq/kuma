package controllers

import (
	"context"
	"fmt"
	"maps"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
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
)

// MeshServiceReconciler reconciles a MeshService object
type MeshServiceReconciler struct {
	kube_client.Client
	kube_record.EventRecorder
	Log    logr.Logger
	Scheme *kube_runtime.Scheme
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

	if len(svc.GetAnnotations()) > 0 {
		if _, ok := svc.GetAnnotations()[metadata.KumaGatewayAnnotation]; ok {
			log.V(1).Info("service is for gateway. Ignoring.")
			return kube_ctrl.Result{}, nil
		}
	}

	if svc.Spec.ClusterIP == "" { // todo(jakubdyszkiewicz) headless service support will come later
		log.V(1).Info("service has no cluster IP. Ignoring.")
		return kube_ctrl.Result{}, nil
	}

	_, ok := svc.GetLabels()[mesh_proto.MeshTag]
	if !ok && !injectedLabel {
		log.V(1).Info("service is not considered to be service in a mesh")
		if err := r.deleteIfExist(ctx, req.NamespacedName); err != nil {
			return kube_ctrl.Result{}, err
		}
		return kube_ctrl.Result{}, nil
	}

	mesh := util.MeshOfByLabelOrAnnotation(log, svc, namespace)

	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: mesh}, &v1alpha1.Mesh{}); err != nil {
		if kube_apierrs.IsNotFound(err) {
			log.V(1).Info("mesh not found")
			if err := r.deleteIfExist(ctx, req.NamespacedName); err != nil {
				return kube_ctrl.Result{}, err
			}
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, err
	}

	if svc.Spec.ClusterIP != kube_core.ClusterIPNone {
		name := kube_client.ObjectKeyFromObject(svc)
		op, err := r.manageMeshService(
			ctx,
			svc,
			mesh,
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
			metadata.KumaSerivceName: svc.Name,
		}),
	); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to list MeshServices for headless Service")
	}
	for _, svc := range meshServices.Items {
		if len(svc.GetOwnerReferences()) == 0 {
			continue
		}
		owner := svc.GetOwnerReferences()[0]
		// TODO check kinds of owner
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

	log.V(1).Info("", "MeshServiceEndpoints", trackedPodEndpoints, "EndpointSliceEndpoints", servicePodEndpoints)

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
		canonicalNameHash := k8s.NewHasher()
		canonicalNameHash.Write([]byte(svc.Name))
		canonicalNameHash.Write([]byte(svc.Namespace))
		canonicalName := fmt.Sprintf("%s-%s", current.Name, k8s.HashToString(canonicalNameHash))
		op, err := r.manageMeshService(
			ctx,
			svc,
			mesh,
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

func (r *MeshServiceReconciler) setFromClusterIPSvc(ms *meshservice_k8s.MeshService, svc *kube_core.Service) error {
	if ms.ObjectMeta.GetGeneration() != 0 {
		if owners := ms.GetOwnerReferences(); len(owners) == 0 || owners[0].UID != svc.GetUID() {
			r.EventRecorder.Eventf(
				svc, kube_core.EventTypeWarning, FailedToGenerateMeshServiceReason, "MeshService already exists and isn't owned by Service",
			)
			return errors.Errorf("MeshService already exists and isn't owned by Service")
		}
	}
	ms.Spec.Selector = meshservice_api.Selector{
		DataplaneTags: svc.Spec.Selector,
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

func (r *MeshServiceReconciler) setFromPodAndHeadlessSvc(endpoint kube_discovery.Endpoint) func(*meshservice_k8s.MeshService, *kube_core.Service) error {
	return func(ms *meshservice_k8s.MeshService, svc *kube_core.Service) error {
		if ms.ObjectMeta.GetGeneration() != 0 {
			if owners := ms.GetOwnerReferences(); len(owners) == 0 || owners[0].UID != endpoint.TargetRef.UID {
				r.EventRecorder.Eventf(
					svc, kube_core.EventTypeWarning, FailedToGenerateMeshServiceReason, "MeshService already exists and isn't owned by Pod",
				)
				return errors.Errorf("MeshService already exists and isn't owned by Pod")
			}
		}
		ms.Spec.Selector = meshservice_api.Selector{
			DataplaneRef: &meshservice_api.DataplaneRef{
				Name: endpoint.TargetRef.Name,
			},
		}
		for _, address := range endpoint.Addresses {
			ms.Status.VIPs = append(ms.Status.VIPs,
				meshservice_api.VIP{
					IP: address,
				})
		}
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
	setSpec func(*meshservice_k8s.MeshService, *kube_core.Service) error,
	meshServiceName kube_types.NamespacedName,
) (kube_controllerutil.OperationResult, error) {
	ms := &meshservice_k8s.MeshService{
		ObjectMeta: v1.ObjectMeta{
			Name:      meshServiceName.Name,
			Namespace: meshServiceName.Namespace,
		},
	}

	return kube_controllerutil.CreateOrUpdate(ctx, r.Client, ms, func() error {
		ms.ObjectMeta.Labels = maps.Clone(svc.GetLabels())
		if ms.ObjectMeta.Labels == nil {
			ms.ObjectMeta.Labels = map[string]string{}
		}
		ms.ObjectMeta.Labels[mesh_proto.MeshTag] = mesh
		ms.ObjectMeta.Labels[metadata.KumaSerivceName] = svc.GetName()
		if ms.Spec == nil {
			ms.Spec = &meshservice_api.MeshService{}
		}

		ms.Spec.Ports = []meshservice_api.Port{}
		for _, port := range svc.Spec.Ports {
			if port.Protocol != kube_core.ProtocolTCP {
				continue
			}
			ms.Spec.Ports = append(ms.Spec.Ports, meshservice_api.Port{
				Port:       uint32(port.Port),
				TargetPort: port.TargetPort,
				Protocol:   core_mesh.Protocol(pointer.DerefOr(port.AppProtocol, "tcp")),
			})
		}

		if ms.Status == nil {
			ms.Status = &meshservice_api.MeshServiceStatus{}
		}
		return setSpec(ms, svc)
	})
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
