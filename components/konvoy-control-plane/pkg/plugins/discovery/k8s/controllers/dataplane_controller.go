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
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	k8s_resources "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

// DataplaneReconciler reconciles a Dataplane object
type DataplaneReconciler struct {
	client.Client
	Converter k8s_resources.Converter
	core_discovery.DiscoverySink
	Log logr.Logger
}

func (r *DataplaneReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("dataplane", req.NamespacedName)

	// Fetch the Dataplane instance
	crd := &mesh_k8s.Dataplane{}
	if err := r.Get(ctx, req.NamespacedName, crd); err != nil {
		if apierrs.IsNotFound(err) {
			return ctrl.Result{}, r.DiscoverySink.OnWorkloadDelete(core_model.ResourceKey{
				Namespace: req.NamespacedName.Namespace,
				Name:      req.NamespacedName.Name,
			})
		}
		log.Error(err, "unable to fetch Dataplane")
		return ctrl.Result{}, err
	}

	dataplane := &mesh_core.DataplaneResource{}
	if err := r.Converter.ToCoreResource(crd, dataplane); err != nil {
		return ctrl.Result{}, err
	} else {
		return ctrl.Result{}, r.DiscoverySink.OnDataplaneUpdate(dataplane)
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
