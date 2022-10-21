package gatewayapi

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"
	gatewayapi_alpha "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/containers"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

// MeshGatewayReconciler reconciles a GatewayAPI MeshGateway object.
type GatewayReconciler struct {
	kube_client.Client
	Log logr.Logger

	Scheme          *kube_runtime.Scheme
	TypeRegistry    k8s_registry.TypeRegistry
	SystemNamespace string
	ProxyFactory    *containers.DataplaneProxyFactory
	ResourceManager manager.ResourceManager
}

// Reconcile handles transforming a gateway-api MeshGateway into a Kuma MeshGateway and
// managing the status of the gateway-api objects.
func (r *GatewayReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	r.Log.V(1).Info("reconcile", "req", req)
	gateway := &gatewayapi.Gateway{}
	if err := r.Get(ctx, req.NamespacedName, gateway); err != nil {
		if kube_apierrs.IsNotFound(err) {
			// We don't know the mesh, but we don't need it to delete our
			// object.
			err := common.ReconcileLabelledObject(ctx, r.Log, r.TypeRegistry, r.Client, req.NamespacedName, core_model.NoMesh, &mesh_proto.MeshGateway{}, nil)
			return kube_ctrl.Result{}, errors.Wrap(err, "could not delete owned MeshGateway.kuma.io")
		}

		return kube_ctrl.Result{}, err
	}

	class, err := common.GetGatewayClass(ctx, r.Client, gateway.Spec.GatewayClassName)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to retrieve GatewayClass referenced by MeshGateway")
	}

	if class == nil || class.Spec.ControllerName != common.ControllerName {
		return kube_ctrl.Result{}, nil
	}

	if !kube_apimeta.IsStatusConditionTrue(class.Status.Conditions, string(gatewayapi.GatewayClassConditionStatusAccepted)) {
		return kube_ctrl.Result{}, nil
	}

	config, err := r.meshGatewayConfigFromClass(ctx, class)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to get Config from GatewayClass")
	}

	ns := kube_core.Namespace{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: gateway.Namespace}, &ns); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to get Namespace of MeshGateway")
	}

	mesh := k8s_util.MeshOf(gateway, &ns)
	gatewaySpec, listenerConditions, err := r.gapiToKumaGateway(ctx, mesh, gateway, config)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "error generating MeshGateway.kuma.io")
	}

	var gatewayInstance *mesh_k8s.MeshGatewayInstance
	if gatewaySpec != nil {
		if err := common.ReconcileLabelledObject(ctx, r.Log, r.TypeRegistry, r.Client, req.NamespacedName, mesh, &mesh_proto.MeshGateway{}, gatewaySpec); err != nil {
			return kube_ctrl.Result{}, errors.Wrap(err, "could not reconcile owned MeshGateway.kuma.io")
		}

		gatewayInstance, err = r.createOrUpdateInstance(ctx, mesh, gateway, config)
		if err != nil {
			return kube_ctrl.Result{}, errors.Wrap(err, "unable to reconcile MeshGatewayInstance")
		}
	}

	if err := r.updateStatus(ctx, gateway, gatewayInstance, listenerConditions); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update MeshGateway status")
	}

	return kube_ctrl.Result{}, nil
}

func (r *GatewayReconciler) meshGatewayConfigFromClass(ctx context.Context, class *gatewayapi.GatewayClass) (mesh_k8s.MeshGatewayConfigSpec, error) {
	ref, _, err := getParametersRef(ctx, r.Client, class.Spec.ParametersRef)
	if err != nil {
		return mesh_k8s.MeshGatewayConfigSpec{}, errors.Wrap(err, "unable to fetch parameters for GatewayClass")
	}

	if ref != nil {
		if ref.Spec.CrossMesh {
			ref.Spec.MeshGatewayCommonConfig.ServiceType = kube_core.ServiceTypeClusterIP
		}

		return ref.Spec, nil
	}

	return mesh_k8s.MeshGatewayConfigSpec{
		MeshGatewayCommonConfig: mesh_k8s.MeshGatewayCommonConfig{
			ServiceType: kube_core.ServiceTypeLoadBalancer,
			Replicas:    1,
		},
	}, nil
}

func (r *GatewayReconciler) createOrUpdateInstance(ctx context.Context, mesh string, gateway *gatewayapi.Gateway, config mesh_k8s.MeshGatewayConfigSpec) (*mesh_k8s.MeshGatewayInstance, error) {
	instance := &mesh_k8s.MeshGatewayInstance{
		ObjectMeta: kube_meta.ObjectMeta{
			Namespace: gateway.Namespace,
			Name:      gateway.Name,
		},
	}

	tags := mesh_proto.Merge(
		common.ServiceTagForGateway(kube_client.ObjectKeyFromObject(gateway)),
		config.Tags,
	)

	if _, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, instance, func() error {
		if instance.Annotations == nil {
			instance.Annotations = map[string]string{}
		}
		instance.Annotations[metadata.KumaMeshAnnotation] = mesh

		instance.Spec = mesh_k8s.MeshGatewayInstanceSpec{
			Tags:                    tags,
			MeshGatewayCommonConfig: config.MeshGatewayCommonConfig,
		}

		err := kube_controllerutil.SetControllerReference(gateway, instance, r.Scheme)
		return errors.Wrap(err, "unable to set MeshGatewayInstance's controller reference to MeshGateway")
	}); err != nil {
		return nil, errors.Wrap(err, "couldn't create MeshGatewayInstance")
	}

	return instance, nil
}

const gatewayIndexField = ".metadata.gateway"

// gatewaysForRoute returns a function that calculates which MeshGateways might
// be affected by changes in an HTTPRoute so they can be reconciled.
func gatewaysForRoute(l logr.Logger) kube_handler.MapFunc {
	l = l.WithName("gatewaysForRoute")

	return func(obj kube_client.Object) []kube_reconcile.Request {
		route, ok := obj.(*gatewayapi.HTTPRoute)
		if !ok {
			l.Error(nil, "unexpected error converting to be mapped %T object to HTTPRoute", obj)
			return nil
		}

		var requests []kube_reconcile.Request
		for _, parentRef := range route.Spec.ParentRefs {
			namespace := route.Namespace
			if parentRef.Namespace != nil {
				namespace = string(*parentRef.Namespace)
			}

			requests = append(
				requests,
				kube_reconcile.Request{
					NamespacedName: kube_types.NamespacedName{Namespace: namespace, Name: string(parentRef.Name)},
				},
			)
		}

		return requests
	}
}

// gatewaysForClass returns a function that calculates which Gateways might
// be affected by changes in a GatewayClass so they can be reconciled.
func gatewaysForClass(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("gatewaysForClass")

	return func(obj kube_client.Object) []kube_reconcile.Request {
		class, ok := obj.(*gatewayapi.GatewayClass)
		if !ok {
			l.Error(nil, "unexpected error converting to be mapped %T object to GatewayClass", obj)
			return nil
		}

		gateways := &gatewayapi.GatewayList{}
		if err := client.List(
			context.Background(), gateways, kube_client.MatchingFields{gatewayClassField: class.Name},
		); err != nil {
			l.Error(err, "unexpected error listing Gateways")
			return nil
		}

		var requests []kube_reconcile.Request
		for _, gateway := range gateways.Items {
			requests = append(requests, kube_reconcile.Request{
				NamespacedName: kube_client.ObjectKeyFromObject(&gateway),
			})
		}

		return requests
	}
}

// gatewaysForConfig returns a function that calculates which Gateways might
// be affected by changes in a MeshGatewayConfig.
func gatewaysForConfig(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("gatewaysForConfig")

	return func(obj kube_client.Object) []kube_reconcile.Request {
		config, ok := obj.(*mesh_k8s.MeshGatewayConfig)
		if !ok {
			l.Error(nil, "unexpected error converting to be mapped %T object to MeshGatewayConfig", obj)
			return nil
		}

		classes := &gatewayapi.GatewayClassList{}
		if err := client.List(
			context.Background(), classes, kube_client.MatchingFields{parametersRefField: config.Name},
		); err != nil {
			l.Error(err, "unexpected error listing GatewayClasses")
			return nil
		}

		var requests []kube_reconcile.Request

		for _, class := range classes.Items {
			gateways := &gatewayapi.GatewayList{}
			if err := client.List(
				context.Background(), gateways, kube_client.MatchingFields{gatewayClassField: class.Name},
			); err != nil {
				l.Error(err, "unexpected error listing Gateways")
				return nil
			}
			for _, gateway := range gateways.Items {
				requests = append(requests, kube_reconcile.Request{
					NamespacedName: kube_client.ObjectKeyFromObject(&gateway),
				})
			}
		}

		return requests
	}
}

// gatewaysForGrant returns a function that calculates which Gateways might
// be affected by changes in a ReferenceGrant so they can be reconciled.
func gatewaysForGrant(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("gatewaysForGrant")

	return func(obj kube_client.Object) []kube_reconcile.Request {
		grant, ok := obj.(*gatewayapi_alpha.ReferenceGrant)
		if !ok {
			l.Error(nil, "unexpected error converting to be mapped %T object to GatewayGrant", obj)
			return nil
		}

		var namespaces []gatewayapi_alpha.Namespace
		for _, from := range grant.Spec.From {
			if from.Group == gatewayapi_alpha.Group(gatewayapi.GroupVersion.Group) && from.Kind == "Gateway" {
				namespaces = append(namespaces, from.Namespace)
			}
		}

		var requests []kube_reconcile.Request

		for _, namespace := range namespaces {
			gateways := &gatewayapi.GatewayList{}
			if err := client.List(context.Background(), gateways, kube_client.InNamespace(namespace)); err != nil {
				l.Error(err, "unexpected error listing Gateways")
				return nil
			}

			for _, gateway := range gateways.Items {
				requests = append(requests, kube_reconcile.Request{
					NamespacedName: kube_client.ObjectKeyFromObject(&gateway),
				})
			}
		}

		return requests
	}
}

func (r *GatewayReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	// This index helps us list routes that point to a MeshGateway in
	// attachedListenersForMeshGateway.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &gatewayapi.HTTPRoute{}, gatewayIndexField, func(obj kube_client.Object) []string {
		route := obj.(*gatewayapi.HTTPRoute)

		var names []string

		for _, parentRef := range route.Spec.ParentRefs {
			namespace := route.Namespace
			if parentRef.Namespace != nil {
				namespace = string(*parentRef.Namespace)
			}

			names = append(
				names,
				kube_types.NamespacedName{Namespace: namespace, Name: string(parentRef.Name)}.String(),
			)
		}

		return names
	}); err != nil {
		return err
	}

	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&gatewayapi.Gateway{}).
		Owns(&mesh_k8s.MeshGateway{}).
		Owns(&mesh_k8s.MeshGatewayInstance{}).
		Watches(
			&kube_source.Kind{Type: &gatewayapi.HTTPRoute{}},
			kube_handler.EnqueueRequestsFromMapFunc(gatewaysForRoute(r.Log)),
		).
		Watches(
			&kube_source.Kind{Type: &gatewayapi.GatewayClass{}},
			kube_handler.EnqueueRequestsFromMapFunc(gatewaysForClass(r.Log, r.Client)),
		).
		Watches(
			&kube_source.Kind{Type: &mesh_k8s.MeshGatewayConfig{}},
			kube_handler.EnqueueRequestsFromMapFunc(gatewaysForConfig(r.Log, r.Client)),
		).
		Watches(
			&kube_source.Kind{Type: &gatewayapi_alpha.ReferenceGrant{}},
			kube_handler.EnqueueRequestsFromMapFunc(gatewaysForGrant(r.Log, r.Client)),
		).
		Complete(r)
}
