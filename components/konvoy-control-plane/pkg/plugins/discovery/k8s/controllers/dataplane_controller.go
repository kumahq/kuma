package controllers

import (
	"context"

	"github.com/go-logr/logr"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	mesh_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

// DataplaneReconciler reconciles a Dataplane object
type DataplaneReconciler struct {
	client.Client
	Log logr.Logger
	core_discovery.DiscoverySink
}

func (r *DataplaneReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("dataplane", req.NamespacedName)

	// Fetch the Dataplane instance
	dataplane := &mesh_k8s.Dataplane{}
	if err := r.Get(ctx, req.NamespacedName, dataplane); err != nil {
		if apierrs.IsNotFound(err) {
			return ctrl.Result{}, r.DiscoverySink.OnWorkloadDelete(req.NamespacedName)
		}
		log.Error(err, "unable to fetch Dataplane")
		return ctrl.Result{}, err
	}

	if wrk, err := DataplaneToWorkload(dataplane); err != nil {
		return ctrl.Result{}, err
	} else {
		return ctrl.Result{}, r.DiscoverySink.OnWorkloadUpdate(wrk)
	}
}

func (r *DataplaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mesh_k8s.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&mesh_k8s.Dataplane{}).
		// on ProxyTemplate update reconcile affected Dataplanes
		Watches(&source.Kind{Type: &mesh_k8s.ProxyTemplate{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: &ProxyTemplateToDataplanesMapper{Client: mgr.GetClient()},
		}).
		Complete(r)
}

type ProxyTemplateToDataplanesMapper struct {
	client.Client
}

func (m *ProxyTemplateToDataplanesMapper) Map(obj handler.MapObject) []reconcile.Request {
	// List Dataplanes in every Namespace
	dataplanes := &mesh_k8s.DataplaneList{}
	if err := m.Client.List(context.Background(), dataplanes); err != nil {
		log := ctrl.Log.WithName("proxytemplate-to-dataplanes-mapper").WithValues("proxytemplate", obj.Meta)
		log.Error(err, "failed to fetch Dataplanes")
		return nil
	}

	var req []reconcile.Request
	for _, dataplane := range dataplanes.Items {
		// TODO(yskopets): match Dataplane against ProxyTemplate's selector

		req = append(req, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: dataplane.Namespace, Name: dataplane.Name},
		})
	}
	return req
}
