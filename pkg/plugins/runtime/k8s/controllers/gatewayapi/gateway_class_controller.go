package gatewayapi

import (
	"context"

	"github.com/go-logr/logr"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"
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

	if class.Spec.ControllerName != controllerName {
		return kube_ctrl.Result{}, nil
	}

	gateways := &gatewayapi.GatewayList{}
	if err := r.Client.List(
		ctx, gateways, kube_client.MatchingFields{gatewayClassKey: class.Name},
	); err != nil {
		return kube_ctrl.Result{}, err
	}

	inUse := len(gateways.Items) > 0

	if controllerutil.ContainsFinalizer(class, gatewayapi.GatewayClassFinalizerGatewaysExist) == inUse {
		return kube_ctrl.Result{}, nil
	}

	if inUse {
		controllerutil.AddFinalizer(class, gatewayapi.GatewayClassFinalizerGatewaysExist)
	} else {
		controllerutil.RemoveFinalizer(class, gatewayapi.GatewayClassFinalizerGatewaysExist)
	}

	if err := r.Update(ctx, class); err != nil {
		return kube_ctrl.Result{}, err
	}

	return kube_ctrl.Result{}, nil
}

// GatewayToClassMapper maps a Gateway object event to a list of GatewayClasses.
func GatewayToClassMapper(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("gateway-to-gateway-class-mapper")

	return func(obj kube_client.Object) []kube_reconcile.Request {
		// If we don't have an object, we need to reconcile all GatewayClasses
		if obj == nil {
			classes := &gatewayapi.GatewayClassList{}
			if err := client.List(context.Background(), classes); err != nil {
				l.Error(err, "failed to fetch GatewayInstances")
			}

			var req []kube_reconcile.Request
			for _, class := range classes.Items {
				if class.Spec.ControllerName != controllerName {
					continue
				}
				req = append(req, kube_reconcile.Request{
					NamespacedName: kube_types.NamespacedName{Namespace: class.Namespace, Name: class.Name},
				})
			}

			return req
		}

		gateway := obj.(*gatewayapi.Gateway)

		return []kube_reconcile.Request{
			{NamespacedName: kube_types.NamespacedName{Name: string(gateway.Spec.GatewayClassName)}},
		}
	}
}

func (r *GatewayClassReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	// Add an index to Gateways for use in Reconcile
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &gatewayapi.Gateway{}, gatewayClassKey, func(obj kube_client.Object) []string {
		instance := obj.(*gatewayapi.Gateway)

		return []string{string(instance.Spec.GatewayClassName)}
	}); err != nil {
		return err
	}

	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&gatewayapi.GatewayClass{}).
		// When something changes with Gateways, we want to reconcile
		// GatewayClasses
		Watches(&kube_source.Kind{Type: &gatewayapi.Gateway{}}, kube_handler.EnqueueRequestsFromMapFunc(GatewayToClassMapper(r.Log, mgr.GetClient()))).
		Complete(r)
}
