package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_events "k8s.io/client-go/tools/events"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	workload_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/api/v1alpha1"
	workload_k8s "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/k8s/v1alpha1"
	mesh_k8s "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
)

const (
	// MultipleMeshesDetectedReason is a Kubernetes event type, used when
	// dataplanes in multiple meshes reference the same workload.
	MultipleMeshesDetectedReason = "MultipleMeshesDetected"
)

// WorkloadReconciler reconciles Workload resources based on Dataplane labels
type WorkloadReconciler struct {
	kube_client.Client
	kube_events.EventRecorder
	Log logr.Logger
}

func (r *WorkloadReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	log := r.Log.WithValues("workload", req.NamespacedName)

	workload := &workload_k8s.Workload{}
	if err := r.Get(ctx, req.NamespacedName, workload); err != nil {
		if !kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, errors.Wrapf(err, "unable to fetch Workload %s", req.Name)
		}
		workload = nil
	}

	dataplanes := &mesh_k8s.DataplaneList{}
	if err := r.List(ctx, dataplanes,
		kube_client.InNamespace(req.Namespace),
		kube_client.MatchingLabels{metadata.KumaWorkload: req.Name},
	); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to list Dataplanes")
	}

	meshName := ""
	for _, dp := range dataplanes.Items {
		if meshName == "" {
			meshName = dp.Mesh
		} else if meshName != dp.Mesh {
			r.handleMultipleMeshesDetected(ctx, req.Namespace, req.Name)
			return kube_ctrl.Result{}, nil
		}
	}
	// If no Dataplanes reference this workload, delete it (if it exists and is managed by us)
	if len(dataplanes.Items) == 0 {
		if workload != nil {
			// Only delete if managed by k8s-controller
			if managedBy, ok := workload.Labels[mesh_proto.ManagedByLabel]; ok && managedBy == "k8s-controller" {
				log.Info("deleting workload with no dataplane references")
				if err := r.Delete(ctx, workload); err != nil && !kube_apierrs.IsNotFound(err) {
					return kube_ctrl.Result{}, errors.Wrapf(err, "failed to delete Workload %s", req.Name)
				}
			}
		}
		return kube_ctrl.Result{}, nil
	}

	if err := r.createOrUpdateWorkload(ctx, req.Name, meshName, req.Namespace); err != nil {
		return kube_ctrl.Result{}, err
	}

	return kube_ctrl.Result{}, nil
}

func (r *WorkloadReconciler) handleMultipleMeshesDetected(ctx context.Context, namespace, workloadName string) {
	log := r.Log.WithValues("workload", workloadName, "namespace", namespace)

	log.Error(errors.New("multiple meshes detected"),
		"namespace has dataplanes in multiple meshes for same workload")
	ns := &kube_core.Namespace{}
	if err := r.Get(ctx, kube_types.NamespacedName{Name: namespace}, ns); err != nil {
		log.V(1).Info("unable to fetch namespace for event emission", "error", err)
	} else {
		r.Eventf(ns, nil, kube_core.EventTypeWarning, MultipleMeshesDetectedReason, "MultipleMeshesDetected",
			"Skipping Workload generation: namespace %s has pods in multiple meshes for workload %s. This configuration is not supported.",
			namespace, workloadName)
	}
}

func (r *WorkloadReconciler) createOrUpdateWorkload(ctx context.Context, workloadName, meshName, namespace string) error {
	log := r.Log.WithValues("workload", workloadName, "mesh", meshName, "namespace", namespace)

	workload := &workload_k8s.Workload{
		ObjectMeta: v1.ObjectMeta{
			Name:      workloadName,
			Namespace: namespace,
		},
	}

	result, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, workload, func() error {
		if workload.Labels == nil {
			workload.Labels = map[string]string{}
		}
		workload.Labels[mesh_proto.MeshTag] = meshName
		workload.Labels[mesh_proto.ManagedByLabel] = "k8s-controller"

		if workload.Spec == nil {
			workload.Spec = &workload_api.Workload{}
		}

		return nil
	})
	if err != nil {
		if kube_apierrs.IsAlreadyExists(err) {
			log.Info("workload already exists")
			return nil
		}
		return errors.Wrapf(err, "failed to create/update Workload %s in namespace %s", workloadName, namespace)
	}

	switch result {
	case kube_controllerutil.OperationResultCreated:
		log.Info("workload created")
	case kube_controllerutil.OperationResultUpdated:
		log.V(1).Info("workload updated")
	}

	return nil
}

func (r *WorkloadReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		Named("kuma-workload-controller").
		For(&workload_k8s.Workload{}).
		Watches(&mesh_k8s.Dataplane{}, kube_handler.EnqueueRequestsFromMapFunc(DataplaneToWorkloadMapper(r.Log))).
		Complete(r)
}

func DataplaneToWorkloadMapper(l logr.Logger) kube_handler.MapFunc {
	l = l.WithName("dataplane-to-workload-mapper")
	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		dp, ok := obj.(*mesh_k8s.Dataplane)
		if !ok {
			l.Error(nil, "unexpected object type", "type", obj.GetObjectKind())
			return nil
		}

		workloadName := dp.GetLabels()[metadata.KumaWorkload]
		if workloadName == "" {
			return nil
		}

		return []kube_reconcile.Request{
			{
				NamespacedName: kube_types.NamespacedName{
					Namespace: dp.Namespace,
					Name:      workloadName,
				},
			},
		}
	}
}
