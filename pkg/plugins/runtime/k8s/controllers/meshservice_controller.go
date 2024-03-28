package controllers

import (
	"context"
	"maps"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
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
	"github.com/kumahq/kuma/pkg/util/pointer"
)

// MeshServiceReconciler reconciles a MeshService object
type MeshServiceReconciler struct {
	kube_client.Client
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

	ms := &meshservice_k8s.MeshService{
		ObjectMeta: v1.ObjectMeta{
			Name:      svc.GetName(),
			Namespace: svc.GetNamespace(),
		},
	}
	if err := kube_controllerutil.SetOwnerReference(svc, ms, r.Scheme); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "could not set owner reference")
	}

	operationResult, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, ms, func() error {
		ms.ObjectMeta.Labels = maps.Clone(svc.GetLabels())
		if ms.ObjectMeta.Labels == nil {
			ms.ObjectMeta.Labels = map[string]string{}
		}
		ms.ObjectMeta.Labels[mesh_proto.MeshTag] = mesh
		if ms.Spec == nil {
			ms.Spec = &meshservice_api.MeshService{}
		}

		ms.Spec.Selector = meshservice_api.Selector{
			DataplaneTags: svc.Spec.Selector,
		}

		ms.Spec.Ports = []meshservice_api.Port{}
		for _, port := range svc.Spec.Ports {
			if port.Protocol != kube_core.ProtocolTCP {
				continue
			}
			ms.Spec.Ports = append(ms.Spec.Ports, meshservice_api.Port{
				Port:       uint32(port.Port),
				TargetPort: uint32(port.TargetPort.IntVal), // todo(jakubdyszkiewicz): update after API changes
				Protocol:   core_mesh.Protocol(pointer.DerefOr(port.AppProtocol, "tcp")),
			})
		}

		ms.Spec.Status.VIPs = []meshservice_api.VIP{
			{
				IP: svc.Spec.ClusterIP,
			},
		}
		return nil
	})
	if err != nil {
		return kube_ctrl.Result{}, err
	}

	log.V(1).Info("mesh service reconciled", "result", operationResult)
	return kube_ctrl.Result{}, nil
}

func (r *MeshServiceReconciler) deleteIfExist(ctx context.Context, key kube_types.NamespacedName) error {
	ms := &meshservice_k8s.MeshService{
		ObjectMeta: v1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
	}
	if err := r.Client.Delete(ctx, ms); err != nil && !kube_apierrs.IsNotFound(err) {
		return errors.Wrap(err, "could not delete mesh service")
	}
	return nil
}

func (r *MeshServiceReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.Service{}).
		// on Namespace update we reconcile Services in this namespace
		Watches(&kube_core.Namespace{}, kube_handler.EnqueueRequestsFromMapFunc(NamespaceToServiceMapper(r.Log, mgr.GetClient())), builder.WithPredicates(predicate.LabelChangedPredicate{})).
		// on Mesh create or delete reconcile all Services
		Watches(&v1alpha1.Mesh{}, kube_handler.EnqueueRequestsFromMapFunc(MeshToMeshService(r.Log, mgr.GetClient())), builder.WithPredicates(CreateOrDeletePredicate{})).
		Complete(r)
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

func MeshToMeshService(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
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
