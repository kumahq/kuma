package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_record "k8s.io/client-go/tools/record"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	workload_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/api/v1alpha1"
	workload_k8s "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/k8s/v1alpha1"
	k8s_common "github.com/kumahq/kuma/v2/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
)

const (
	// CreatedWorkloadReason is added to an event when
	// a new Workload is successfully created.
	CreatedWorkloadReason = "CreatedWorkload"
	// UpdatedWorkloadReason is added to an event when
	// an existing Workload is successfully updated.
	UpdatedWorkloadReason = "UpdatedWorkload"
	// DeletedWorkloadReason is added to an event when
	// a Workload is successfully deleted.
	DeletedWorkloadReason = "DeletedWorkload"
)

// WorkloadReconciler reconciles Workload resources based on Dataplane labels
type WorkloadReconciler struct {
	kube_client.Client
	kube_record.EventRecorder
	Log               logr.Logger
	Scheme            *kube_runtime.Scheme
	ResourceConverter k8s_common.Converter
}

func (r *WorkloadReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	log := r.Log.WithValues("workload", req.NamespacedName)

	// Fetch the Workload
	workload := &workload_k8s.Workload{}
	if err := r.Get(ctx, req.NamespacedName, workload); err != nil {
		if !kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, errors.Wrapf(err, "unable to fetch Workload %s", req.Name)
		}
		// Workload doesn't exist, we may need to create it
		workload = nil
	}

	// List all Dataplanes in the namespace that reference this workload
	dataplanes := &mesh_k8s.DataplaneList{}
	if err := r.List(ctx, dataplanes, kube_client.InNamespace(req.Namespace)); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to list Dataplanes")
	}

	// Find Dataplanes that reference this workload
	var referencingDPs []mesh_k8s.Dataplane
	var meshName string
	for _, dp := range dataplanes.Items {
		if workloadName, ok := dp.GetAnnotations()[metadata.KumaWorkload]; ok && workloadName == req.Name {
			referencingDPs = append(referencingDPs, dp)
			if meshName == "" {
				meshName = dp.Mesh
			}
		}
	}

	// If no Dataplanes reference this workload, delete it (if it exists and is managed by us)
	if len(referencingDPs) == 0 {
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

	// Create or update the Workload
	if err := r.createOrUpdateWorkload(ctx, req.Name, meshName, req.Namespace); err != nil {
		return kube_ctrl.Result{}, err
	}

	return kube_ctrl.Result{}, nil
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
		// Set labels
		if workload.Labels == nil {
			workload.Labels = map[string]string{}
		}
		workload.Labels[mesh_proto.MeshTag] = meshName
		workload.Labels[mesh_proto.ManagedByLabel] = "k8s-controller"

		// Ensure spec is initialized (empty struct is fine)
		if workload.Spec == nil {
			workload.Spec = &workload_api.Workload{}
		}

		// Don't set status - leave it nil
		return nil
	})
	if err != nil {
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

		workloadName, ok := dp.GetAnnotations()[metadata.KumaWorkload]
		if !ok || workloadName == "" {
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
