package gatewayapi

import (
	"context"

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
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
)

// GatewayClassReconciler reconciles a GatewayAPI GatewayClass object.
type GatewayClassReconciler struct {
	kube_client.Client
	Log logr.Logger
}

const gatewayClassKey = ".metadata.gatewayClass"

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
		ctx, gateways, kube_client.MatchingFields{gatewayClassKey: class.Name},
	); err != nil {
		return kube_ctrl.Result{}, err
	}

	if len(gateways.Items) > 0 {
		controllerutil.AddFinalizer(class, gatewayapi.GatewayClassFinalizerGatewaysExist)
	} else {
		controllerutil.RemoveFinalizer(class, gatewayapi.GatewayClassFinalizerGatewaysExist)
	}

	if err := r.Update(ctx, class); err != nil {
		return kube_ctrl.Result{}, err
	}

	// Prepare modified object for patching status
	updated := class.DeepCopy()

	kube_apimeta.SetStatusCondition(
		&updated.Status.Conditions,
		kube_meta.Condition{
			Type:               string(gatewayapi.GatewayClassConditionStatusAccepted),
			Status:             kube_meta.ConditionTrue,
			Reason:             string(gatewayapi.GatewayClassReasonAccepted),
			ObservedGeneration: class.GetGeneration(),
		},
	)

	if err := r.Client.Status().Patch(ctx, updated, kube_client.MergeFrom(class)); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update status subresource")
	}

	return kube_ctrl.Result{}, nil
}

// gatewayToClassMapper maps a Gateway object event to a list of GatewayClasses.
func gatewayToClassMapper(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("gatewayToClassMapper")

	return func(obj kube_client.Object) []kube_reconcile.Request {
		// If we don't have an object, we need to reconcile all GatewayClasses
		if obj == nil {
			classes := &gatewayapi.GatewayClassList{}
			if err := client.List(context.Background(), classes); err != nil {
				l.Error(err, "failed to list GatewayClasses")
			}

			var req []kube_reconcile.Request
			for _, class := range classes.Items {
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

	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(), &gatewayapi.Gateway{}, gatewayClassKey, gatewayClassNameIndexer,
	); err != nil {
		return err
	}

	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&gatewayapi.GatewayClass{}).
		// When something changes with Gateways, we want to reconcile
		// GatewayClasses
		Watches(&kube_source.Kind{Type: &gatewayapi.Gateway{}}, kube_handler.EnqueueRequestsFromMapFunc(gatewayToClassMapper(r.Log, r.Client))).
		Complete(r)
}
