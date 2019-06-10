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

	corev1 "k8s.io/api/core/v1"
)

type PodObserver interface {
	OnUpdate(*corev1.Pod) error
	OnDelete(types.NamespacedName) error
}

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Log      logr.Logger
	Observer PodObserver
}

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

func (r *PodReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pod", req.NamespacedName)

	if r.Observer == nil {
		return ctrl.Result{}, nil
	}

	// Fetch the Pod instance
	pod := &corev1.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		log.Error(err, "unable to fetch Pod")
		if apierrs.IsNotFound(err) {
			return ctrl.Result{}, r.Observer.OnDelete(req.NamespacedName)
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.Observer.OnUpdate(pod)
}

func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := corev1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
