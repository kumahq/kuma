/*
Copyright 2019 Konvoy authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	kube_core "k8s.io/api/core/v1"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Log logr.Logger
	core_discovery.DiscoverySink
}

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

func (r *PodReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pod", req.NamespacedName)

	// Fetch the Pod instance
	pod := &kube_core.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		log.Error(err, "unable to fetch Pod")
		if apierrs.IsNotFound(err) {
			return ctrl.Result{}, r.DiscoverySink.OnWorkloadDelete(req.NamespacedName)
		}
		return ctrl.Result{}, err
	}

	if wrk, err := ToWorkload(pod); err != nil {
		return ctrl.Result{}, err
	} else {
		return ctrl.Result{}, r.DiscoverySink.OnWorkloadUpdate(wrk)
	}
}

func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := kube_core.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.Pod{}).
		// on ProxyTemplate update reconcile affected Pods
		Watches(&source.Kind{Type: &mesh_k8s.ProxyTemplate{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: &ProxyTemplateToPodsMapper{Client: mgr.GetClient()},
		}).
		Complete(r)
}

type ProxyTemplateToPodsMapper struct {
	client.Client
}

func (m *ProxyTemplateToPodsMapper) Map(tmpl handler.MapObject) []reconcile.Request {
	// List all Pods in the same Namespace
	pods := &kube_core.PodList{}
	if err := m.Client.List(context.Background(), pods, client.InNamespace(tmpl.Meta.GetNamespace())); err != nil {
		log := ctrl.Log.WithName("proxytemplate-to-pods-mapper").WithValues("proxytemplate", tmpl.Meta)
		log.Error(err, "failed to fetch Pods", "namespace", tmpl.Meta.GetNamespace())
		return nil
	}

	var req []reconcile.Request
	for i := range pods.Items {
		pod := &pods.Items[i]
		if pod.GetAnnotations() != nil && pod.GetAnnotations()[mesh_k8s.ProxyTemplateAnnotation] == tmpl.Meta.GetName() {
			req = append(req, reconcile.Request{
				NamespacedName: types.NamespacedName{Namespace: pod.GetNamespace(), Name: pod.GetName()},
			})
		}
	}
	return req
}
