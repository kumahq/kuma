package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_discovery "k8s.io/api/discovery/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_event "k8s.io/client-go/tools/events"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	meshzoneaddress_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshzoneaddress/api/v1alpha1"
	meshzoneaddress_k8s "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshzoneaddress/k8s/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/util"
)

const (
	// KumaZoneProxyTypeIngress marks a Service as the public endpoint for a
	// mesh-scoped zone ingress proxy.
	KumaZoneProxyTypeIngress          = "ingress"
	CreatedMeshZoneAddressReason      = "CreatedMeshZoneAddress"
	UpdatedMeshZoneAddressReason      = "UpdatedMeshZoneAddress"
	NoPublicAddressForZoneProxyReason = "NoPublicAddress"
)

// MeshZoneAddressReconciler watches Services labeled with
// k8s.kuma.io/zone-proxy-type=ingress and maintains a MeshZoneAddress
// resource holding the public address and port for cross-zone routing.
type MeshZoneAddressReconciler struct {
	kube_client.Client
	kube_event.EventRecorder
	Log      logr.Logger
	Scheme   *kube_runtime.Scheme
	ZoneName string
}

func (r *MeshZoneAddressReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	log := r.Log.WithValues("service", req.NamespacedName)

	namespace := &kube_core.Namespace{}
	if err := r.Get(ctx, kube_types.NamespacedName{Name: req.Namespace}, namespace); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to get Namespace for Service")
	}

	svc := &kube_core.Service{}
	if err := r.Get(ctx, req.NamespacedName, svc); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, errors.Wrapf(err, "unable to fetch Service %s", req.Name)
	}

	// Only handle Services labeled as zone-proxy ingress.
	if svc.GetLabels()[metadata.KumaZoneProxyTypeLabel] != KumaZoneProxyTypeIngress {
		return kube_ctrl.Result{}, r.deleteIfExists(ctx, req.NamespacedName)
	}

	// Require at least one ready endpoint before publishing the address.
	ready, err := r.hasReadyEndpoints(ctx, svc)
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	if !ready {
		log.V(1).Info("no ready endpoints, removing MeshZoneAddress")
		return kube_ctrl.Result{}, r.deleteIfExists(ctx, req.NamespacedName)
	}

	// Resolve public address and port from the Service.
	address, port, err := r.resolveCoordinates(ctx, log, svc)
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	if address == "" {
		r.Eventf(svc, nil, kube_core.EventTypeWarning, NoPublicAddressForZoneProxyReason, "NoPublicAddress",
			"unable to determine public address for zone ingress Service; ensure it exposes a reachable external address (LoadBalancer, NodePort with suitable node addresses, or spec.externalIPs) and that the address is ready")
		return kube_ctrl.Result{}, r.deleteIfExists(ctx, req.NamespacedName)
	}

	meshName := util.MeshOfByLabelOrAnnotation(log, svc, namespace)

	mza := &meshzoneaddress_k8s.MeshZoneAddress{
		ObjectMeta: v1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}

	result, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, mza, func() error {
		// If the MeshZoneAddress already exists and is not owned by this Service,
		// skip mutation to avoid clobbering user-managed resources.
		if mza.GetGeneration() != 0 {
			if owners := mza.GetOwnerReferences(); len(owners) == 0 || owners[0].UID != svc.GetUID() {
				r.Eventf(svc, nil, kube_core.EventTypeWarning, NoPublicAddressForZoneProxyReason, "Conflict",
					"MeshZoneAddress %s already exists and is not owned by this Service", req.Name)
				return errors.Errorf("MeshZoneAddress already exists and is not owned by Service")
			}
		}
		if mza.Labels == nil {
			mza.Labels = map[string]string{}
		}
		mza.Labels[mesh_proto.MeshTag] = meshName
		mza.Labels[mesh_proto.ZoneTag] = r.ZoneName
		mza.Labels[mesh_proto.ManagedByLabel] = "k8s-controller"
		mza.Labels[mesh_proto.EnvTag] = mesh_proto.KubernetesEnvironment

		if mza.Spec == nil {
			mza.Spec = &meshzoneaddress_api.MeshZoneAddress{}
		}
		mza.Spec.Address = address
		mza.Spec.Port = port

		return kube_controllerutil.SetOwnerReference(svc, mza, r.Scheme)
	})
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to create or update MeshZoneAddress")
	}

	switch result {
	case kube_controllerutil.OperationResultCreated:
		r.Eventf(svc, nil, kube_core.EventTypeNormal, CreatedMeshZoneAddressReason, "Create",
			"Created MeshZoneAddress %s", req.Name)
	case kube_controllerutil.OperationResultUpdated:
		r.Eventf(svc, nil, kube_core.EventTypeNormal, UpdatedMeshZoneAddressReason, "Update",
			"Updated MeshZoneAddress %s", req.Name)
	}

	return kube_ctrl.Result{}, nil
}

// resolveCoordinates determines the public address and port for the Service.
// Priority: externalIPs[0] → LoadBalancer (hostname > IP) → NodePort.
// Returns ("", 0, nil) for unsupported Service types as callers emit a warning.
func (r *MeshZoneAddressReconciler) resolveCoordinates(
	ctx context.Context,
	log logr.Logger,
	svc *kube_core.Service,
) (string, int32, error) {
	if len(svc.Spec.ExternalIPs) > 0 && len(svc.Spec.Ports) > 0 {
		return svc.Spec.ExternalIPs[0], svc.Spec.Ports[0].Port, nil
	}
	switch svc.Spec.Type {
	case kube_core.ServiceTypeLoadBalancer:
		return r.coordinatesFromLoadBalancer(log, svc)
	case kube_core.ServiceTypeNodePort:
		return r.coordinatesFromNodePort(ctx, svc)
	default:
		return "", 0, nil
	}
}

func (r *MeshZoneAddressReconciler) coordinatesFromLoadBalancer(
	log logr.Logger,
	svc *kube_core.Service,
) (string, int32, error) {
	if len(svc.Status.LoadBalancer.Ingress) == 0 || len(svc.Spec.Ports) == 0 {
		log.V(1).Info("LoadBalancer not yet ready")
		return "", 0, nil
	}
	// Hostname takes precedence over IP for stability (MADR-096).
	ingress := svc.Status.LoadBalancer.Ingress[0]
	address := ingress.Hostname
	if address == "" {
		address = ingress.IP
	}
	if address == "" {
		log.V(1).Info("LoadBalancer ingress has neither hostname nor IP")
		return "", 0, nil
	}
	return address, svc.Spec.Ports[0].Port, nil
}

func (r *MeshZoneAddressReconciler) coordinatesFromNodePort(
	ctx context.Context,
	svc *kube_core.Service,
) (string, int32, error) {
	if len(svc.Spec.Ports) == 0 {
		return "", 0, nil
	}
	nodes := &kube_core.NodeList{}
	if err := r.List(ctx, nodes); err != nil {
		return "", 0, errors.Wrap(err, "unable to list Nodes")
	}
	if len(nodes.Items) == 0 {
		return "", 0, errors.New("no nodes found")
	}
	for _, addrType := range NodePortAddressPriority {
		for _, addr := range nodes.Items[0].Status.Addresses {
			if addr.Type == addrType {
				return addr.Address, svc.Spec.Ports[0].NodePort, nil
			}
		}
	}
	return "", 0, nil
}

func (r *MeshZoneAddressReconciler) hasReadyEndpoints(ctx context.Context, svc *kube_core.Service) (bool, error) {
	slices := &kube_discovery.EndpointSliceList{}
	if err := r.List(ctx, slices,
		kube_client.InNamespace(svc.Namespace),
		kube_client.MatchingLabels{kube_discovery.LabelServiceName: svc.Name},
	); err != nil {
		return false, errors.Wrap(err, "unable to list EndpointSlices")
	}
	for i := range slices.Items {
		for j := range slices.Items[i].Endpoints {
			ep := &slices.Items[i].Endpoints[j]
			if ep.Conditions.Ready != nil && *ep.Conditions.Ready {
				return true, nil
			}
		}
	}
	return false, nil
}

func (r *MeshZoneAddressReconciler) deleteIfExists(ctx context.Context, key kube_types.NamespacedName) error {
	mza := &meshzoneaddress_k8s.MeshZoneAddress{
		ObjectMeta: v1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
	}
	if err := r.Delete(ctx, mza); err != nil && !kube_apierrs.IsNotFound(err) {
		return errors.Wrap(err, "unable to delete MeshZoneAddress")
	}
	return nil
}

const zoneProxyTypeLabelIndex = "metadata.labels.zone-proxy-type"

func (r *MeshZoneAddressReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(), &kube_core.Service{}, zoneProxyTypeLabelIndex,
		func(obj kube_client.Object) []string {
			if v := obj.GetLabels()[metadata.KumaZoneProxyTypeLabel]; v != "" {
				return []string{v}
			}
			return nil
		},
	); err != nil {
		return errors.Wrap(err, "failed to index Service by zone-proxy-type label")
	}
	return kube_ctrl.NewControllerManagedBy(mgr).
		Named("kuma-mesh-zone-address-controller").
		For(&kube_core.Service{}).
		Watches(
			&kube_discovery.EndpointSlice{},
			kube_handler.EnqueueRequestsFromMapFunc(EndpointSliceToServicesMapper(r.Log, mgr.GetClient())),
		).
		Watches(
			&kube_core.Node{},
			kube_handler.EnqueueRequestsFromMapFunc(r.nodeToZoneProxyServices(mgr.GetClient())),
		).
		Watches(
			&kube_core.Namespace{},
			kube_handler.EnqueueRequestsFromMapFunc(NamespaceToServiceMapper(r.Log, mgr.GetClient())),
			builder.WithPredicates(predicate.LabelChangedPredicate{}),
		).
		Complete(r)
}

// nodeToZoneProxyServices re-queues all zone-proxy ingress Services when a
// Node changes (needed for NodePort address resolution).
func (r *MeshZoneAddressReconciler) nodeToZoneProxyServices(c kube_client.Client) kube_handler.MapFunc {
	return func(ctx context.Context, _ kube_client.Object) []kube_ctrl.Request {
		svcs := &kube_core.ServiceList{}
		if err := c.List(ctx, svcs, kube_client.MatchingFields{
			zoneProxyTypeLabelIndex: KumaZoneProxyTypeIngress,
		}); err != nil {
			r.Log.Error(err, "failed to list zone-proxy Services on node event")
			return nil
		}
		reqs := make([]kube_ctrl.Request, 0, len(svcs.Items))
		for _, svc := range svcs.Items {
			reqs = append(reqs, kube_ctrl.Request{
				NamespacedName: kube_types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name},
			})
		}
		return reqs
	}
}
