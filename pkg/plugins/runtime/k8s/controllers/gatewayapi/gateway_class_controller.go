package gatewayapi

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
)

// GatewayClassReconciler reconciles a GatewayAPI GatewayClass object.
type GatewayClassReconciler struct {
	kube_client.Client
	Log logr.Logger
}

// gatewayClassField is needed for both GatewayClassReconciler and
// GatewayReconciler.
const gatewayClassField = ".metadata.gatewayClass"

// parametersRefField is important for both GatewayReconciler and
// GatewayClassReconciler.
const parametersRefField = ".metadata.parametersRef"

// Reconcile handles maintaining the GatewayClass finalizer.
func (r *GatewayClassReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	class := &gatewayapi.GatewayClass{}
	if err := r.Get(ctx, req.NamespacedName, class); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}

		return kube_ctrl.Result{}, err
	}

	if class.Spec.ControllerName != common.ControllerName {
		return kube_ctrl.Result{}, nil
	}

	gateways := &gatewayapi.GatewayList{}
	if err := r.Client.List(
		ctx, gateways, kube_client.MatchingFields{gatewayClassField: class.Name},
	); err != nil {
		return kube_ctrl.Result{}, err
	}

	if len(gateways.Items) > 0 {
		controllerutil.AddFinalizer(class, gatewayapi_v1.GatewayClassFinalizerGatewaysExist)
	} else {
		controllerutil.RemoveFinalizer(class, gatewayapi_v1.GatewayClassFinalizerGatewaysExist)
	}

	if err := r.Update(ctx, class); err != nil {
		return kube_ctrl.Result{}, err
	}

	// Prepare modified object for patching status
	updated := class.DeepCopy()

	_, condition, err := getParametersRef(ctx, r.Client, class.Spec.ParametersRef)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to get parametersRef")
	}

	if condition == nil {
		condition = &kube_meta.Condition{
			Type:   string(gatewayapi_v1.GatewayClassConditionStatusAccepted),
			Status: kube_meta.ConditionTrue,
			Reason: string(gatewayapi_v1.GatewayClassReasonAccepted),
		}
	}

	condition.ObservedGeneration = class.GetGeneration()

	kube_apimeta.SetStatusCondition(
		&updated.Status.Conditions,
		*condition,
	)

	if err := r.Client.Status().Patch(ctx, updated, kube_client.MergeFrom(class)); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update status subresource")
	}

	return kube_ctrl.Result{}, nil
}

func getParametersRef(
	ctx context.Context,
	client kube_client.Client,
	parametersRef *gatewayapi.ParametersReference,
) (*mesh_k8s.MeshGatewayConfig, *kube_meta.Condition, error) {
	if parametersRef == nil {
		return nil, nil, nil
	}

	condition := kube_meta.Condition{
		Type:   string(gatewayapi_v1.GatewayClassConditionStatusAccepted),
		Status: kube_meta.ConditionFalse,
		Reason: string(gatewayapi_v1.GatewayClassReasonInvalidParameters),
	}

	if parametersRef.Group != gatewayapi.Group(mesh_k8s.GroupVersion.Group) || parametersRef.Kind != "MeshGatewayConfig" {
		condition.Message = fmt.Sprintf("parametersRef must point to a MeshGatewayConfig.%s", mesh_k8s.GroupVersion.Group)
		return nil, &condition, nil
	}

	if parametersRef.Namespace != nil && *parametersRef.Namespace != "" {
		condition.Message = "parametersRef must not refer to a namespace"
		return nil, &condition, nil
	}

	namespacedName := kube_types.NamespacedName{Name: parametersRef.Name}

	config := &mesh_k8s.MeshGatewayConfig{}
	if err := client.Get(ctx, namespacedName, config); err != nil {
		if kube_apierrs.IsNotFound(err) {
			condition.Message = "parametersRef could not be found"
			return nil, &condition, nil
		}

		return nil, nil, errors.Wrapf(err, "unable to get MeshGatewayConfig %s", namespacedName.String())
	}

	return config, nil, nil
}

// gatewayToClassMapper maps a Gateway object event to a list of GatewayClasses.
func gatewayToClassMapper(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("gatewayToClassMapper")

	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		// If we don't have an object, we need to reconcile all GatewayClasses
		if obj == nil {
			classes := &gatewayapi.GatewayClassList{}
			if err := client.List(ctx, classes); err != nil {
				l.Error(err, "failed to list GatewayClasses")
			}

			var req []kube_reconcile.Request
			for i := range classes.Items {
				class := classes.Items[i]
				if class.Spec.ControllerName != common.ControllerName {
					continue
				}

				req = append(req, kube_reconcile.Request{
					NamespacedName: kube_client.ObjectKeyFromObject(&class),
				})
			}

			return req
		}

		gateway, ok := obj.(*gatewayapi.Gateway)
		if !ok {
			l.Error(nil, "could not convert to Gateway", "object", obj)
		}

		return []kube_reconcile.Request{
			{NamespacedName: kube_types.NamespacedName{Name: string(gateway.Spec.GatewayClassName)}},
		}
	}
}

// gatewayClassesForConfig returns a function that calculates which
// GatewayClasses might be affected by changes in a MeshGatewayConfig.
func gatewayClassesForConfig(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("gatewaysForConfig")

	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		config, ok := obj.(*mesh_k8s.MeshGatewayConfig)
		if !ok {
			l.Error(nil, "unexpected error converting to be mapped object to MeshGatewayConfig", "typ", reflect.TypeOf(obj))
			return nil
		}

		classes := &gatewayapi.GatewayClassList{}
		if err := client.List(
			ctx, classes, kube_client.MatchingFields{parametersRefField: config.Name},
		); err != nil {
			l.Error(err, "unexpected error listing GatewayClasses")
			return nil
		}

		var requests []kube_reconcile.Request

		for i := range classes.Items {
			class := classes.Items[i]
			requests = append(requests, kube_reconcile.Request{
				NamespacedName: kube_client.ObjectKeyFromObject(&class),
			})
		}

		return requests
	}
}

func (r *GatewayClassReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	// Add an index to Gateways for use in Reconcile
	indexLog := r.Log.WithName("gatewayClassNameIndexer")

	gatewayClassNameIndexer := func(obj kube_client.Object) []string {
		gateway, ok := obj.(*gatewayapi.Gateway)
		if !ok {
			indexLog.Error(nil, "could not convert to Gateway", "object", obj)
			return []string{}
		}

		return []string{string(gateway.Spec.GatewayClassName)}
	}

	// this index is also needed by GatewayReconciler!
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(), &gatewayapi.Gateway{}, gatewayClassField, gatewayClassNameIndexer,
	); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &gatewayapi.GatewayClass{}, parametersRefField, func(obj kube_client.Object) []string {
		class := obj.(*gatewayapi.GatewayClass)

		ref := class.Spec.ParametersRef

		if class.Spec.ControllerName != common.ControllerName || ref == nil || ref.Kind != "MeshGatewayConfig" || ref.Group != gatewayapi.Group(mesh_k8s.GroupVersion.Group) {
			return nil
		}

		return []string{ref.Name}
	}); err != nil {
		return err
	}

	return kube_ctrl.NewControllerManagedBy(mgr).
		Named("kuma-gateway-class-controller").
		For(&gatewayapi.GatewayClass{}).
		// When something changes with Gateways, we want to reconcile
		// GatewayClasses
		Watches(
			&gatewayapi.Gateway{},
			kube_handler.EnqueueRequestsFromMapFunc(gatewayToClassMapper(r.Log, r.Client)),
		).
		Watches(
			&mesh_k8s.MeshGatewayConfig{},
			kube_handler.EnqueueRequestsFromMapFunc(gatewayClassesForConfig(r.Log, r.Client)),
		).
		Complete(r)
}
