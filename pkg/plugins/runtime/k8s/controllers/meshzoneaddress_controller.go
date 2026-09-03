package controllers

import (
	"context"
	"slices"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_discovery "k8s.io/api/discovery/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_event "k8s.io/client-go/tools/events"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	meshzoneaddress_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshzoneaddress/api/v1alpha1"
	meshzoneaddress_k8s "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshzoneaddress/k8s/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/util"
)

const (
	CreatedMeshZoneAddressReason      = "CreatedMeshZoneAddress"
	UpdatedMeshZoneAddressReason      = "UpdatedMeshZoneAddress"
	NoPublicAddressForZoneProxyReason = "NoPublicAddress"
)

// List of priority for picking IP when Service that selects a zone proxy is of type NodePort
// We first try to find ExternalIP and then InternalIP.
// ExternalIP will be available in public clouds like GCP, but not on Kind or Minikube.
// On the other hand, on Kind with multizone, there is a connectivity between clusters using InternalIP.
// Technically there is a risk that we will pick InternalIP and other cluster will try to access it without connectivity between them.
// However, in most cases, LoadBalancer will be used anyways, therefore we accept this risk.
var NodePortAddressPriority = []kube_core.NodeAddressType{
	kube_core.NodeExternalIP,
	kube_core.NodeInternalIP,
}

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
	endpointNodes, ready, err := r.readyEndpointNodes(ctx, svc)
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	if !ready {
		log.V(1).Info("no ready endpoints, removing MeshZoneAddress")
		return kube_ctrl.Result{}, r.deleteIfExists(ctx, req.NamespacedName)
	}

	// Resolve public address and port from the Service.
	address, port, err := r.resolveCoordinates(ctx, log, svc, endpointNodes)
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	if address == "" {
		r.Eventf(svc, nil, kube_core.EventTypeWarning, NoPublicAddressForZoneProxyReason, "NoPublicAddress",
			"unable to determine public address for zone ingress Service; ensure it exposes a reachable external address (LoadBalancer, NodePort with suitable node addresses, or spec.externalIPs) and that the address is ready")
		return kube_ctrl.Result{}, r.deleteIfExists(ctx, req.NamespacedName)
	}

	meshName := util.MeshOfByLabel(svc, namespace)

	mza := &meshzoneaddress_k8s.MeshZoneAddress{
		Name:      req.Name,
		Namespace: req.Namespace,
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
	endpointNodes map[string]struct{},
) (string, int32, error) {
	if len(svc.Spec.ExternalIPs) > 0 && len(svc.Spec.Ports) > 0 {
		return svc.Spec.ExternalIPs[0], svc.Spec.Ports[0].Port, nil
	}
	switch svc.Spec.Type {
	case kube_core.ServiceTypeLoadBalancer:
		return r.coordinatesFromLoadBalancer(log, svc)
	case kube_core.ServiceTypeNodePort:
		return r.coordinatesFromNodePort(ctx, log, svc, endpointNodes)
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
	log logr.Logger,
	svc *kube_core.Service,
	endpointNodes map[string]struct{},
) (string, int32, error) {
	if len(svc.Spec.Ports) == 0 {
		return "", 0, nil
	}
	nodes := &kube_core.NodeList{}
	if err := r.List(ctx, nodes); err != nil {
		return "", 0, errors.Wrap(err, "unable to list Nodes")
	}
	candidates := candidateNodes(nodes.Items, endpointNodes, svc.Spec.ExternalTrafficPolicy == kube_core.ServiceExternalTrafficPolicyLocal)
	if len(candidates) == 0 {
		log.V(1).Info("no node able to serve the zone proxy NodePort")
		return "", 0, nil
	}
	for _, addrType := range NodePortAddressPriority {
		for _, node := range candidates {
			for _, addr := range node.Status.Addresses {
				if addr.Type == addrType {
					return addr.Address, svc.Spec.Ports[0].NodePort, nil
				}
			}
		}
	}
	return "", 0, nil
}

// candidateNodes returns the Nodes that can be advertised for a NodePort Service,
// ordered by preference. Nodes hosting a ready endpoint come first, so the advertised
// node is one that actually serves the zone proxy, followed by any other Ready and
// uncordoned node, which kube-proxy still routes from. Each group is sorted by name so
// that the pick is stable across reconciles instead of following the cache iteration
// order.
//
// The caller walks the candidates once per address type, so a reachable ExternalIP on a
// non-serving node wins over a serving node that only has an InternalIP.
//
// Under externalTrafficPolicy: Local only a node with a local endpoint serves the
// traffic, so the fallback group is dropped - unless we do not know where the endpoints
// run, since nodeName is optional in the EndpointSlice API and publishing nothing at all
// would be worse than kube-proxy's own guess.
func candidateNodes(nodes []kube_core.Node, endpointNodes map[string]struct{}, localTrafficPolicy bool) []*kube_core.Node {
	var serving, healthy []*kube_core.Node
	for i := range nodes {
		node := &nodes[i]
		if !isNodeReady(node) {
			continue
		}
		if _, ok := endpointNodes[node.Name]; ok {
			// A cordoned node is still serving the endpoint it already runs.
			serving = append(serving, node)
		} else if !node.Spec.Unschedulable {
			healthy = append(healthy, node)
		}
	}
	byName := func(a, b *kube_core.Node) int { return strings.Compare(a.Name, b.Name) }
	slices.SortFunc(serving, byName)
	slices.SortFunc(healthy, byName)
	if localTrafficPolicy && len(endpointNodes) > 0 {
		return serving
	}
	return append(serving, healthy...)
}

func isNodeReady(node *kube_core.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == kube_core.NodeReady {
			return cond.Status == kube_core.ConditionTrue
		}
	}
	return false
}

// readyEndpointNodes returns the names of the Nodes hosting a ready endpoint of the
// Service, and whether the Service has any ready endpoint at all. Both are needed:
// nodeName is optional in the EndpointSlice API, so the set can be empty even for a
// Service that is perfectly ready.
func (r *MeshZoneAddressReconciler) readyEndpointNodes(ctx context.Context, svc *kube_core.Service) (map[string]struct{}, bool, error) {
	epSlices := &kube_discovery.EndpointSliceList{}
	if err := r.List(ctx, epSlices,
		kube_client.InNamespace(svc.Namespace),
		kube_client.MatchingLabels{kube_discovery.LabelServiceName: svc.Name},
	); err != nil {
		return nil, false, errors.Wrap(err, "unable to list EndpointSlices")
	}
	nodes := map[string]struct{}{}
	ready := false
	for i := range epSlices.Items {
		for j := range epSlices.Items[i].Endpoints {
			ep := &epSlices.Items[i].Endpoints[j]
			if ep.Conditions.Ready == nil || !*ep.Conditions.Ready {
				continue
			}
			ready = true
			if ep.NodeName != nil && *ep.NodeName != "" {
				nodes[*ep.NodeName] = struct{}{}
			}
		}
	}
	return nodes, ready, nil
}

func (r *MeshZoneAddressReconciler) deleteIfExists(ctx context.Context, key kube_types.NamespacedName) error {
	mza := &meshzoneaddress_k8s.MeshZoneAddress{
		Name:      key.Name,
		Namespace: key.Namespace,
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

// nodeToZoneProxyServices re-queues NodePort zone-proxy ingress Services when a
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
			if svc.Spec.Type != kube_core.ServiceTypeNodePort {
				continue
			}
			reqs = append(reqs, kube_ctrl.Request{
				Namespace: svc.Namespace, Name: svc.Name,
			})
		}
		return reqs
	}
}
