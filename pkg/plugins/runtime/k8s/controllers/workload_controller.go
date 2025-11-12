package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_record "k8s.io/client-go/tools/record"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
	log := r.Log.WithValues("dataplane", req.NamespacedName)

	// Fetch the Dataplane
	dp := &mesh_k8s.Dataplane{}
	if err := r.Get(ctx, req.NamespacedName, dp); err != nil {
		if kube_apierrs.IsNotFound(err) {
			// Dataplane deleted, check if we should cleanup Workload
			log.V(1).Info("dataplane not found, checking for orphaned workloads")
			return kube_ctrl.Result{}, r.cleanupOrphanedWorkloads(ctx, req.Namespace)
		}
		return kube_ctrl.Result{}, errors.Wrapf(err, "unable to fetch Dataplane %s", req.Name)
	}

	// Extract workload label from Dataplane
	workloadName, ok := dp.GetLabels()[metadata.KumaWorkload]
	if !ok || workloadName == "" {
		log.V(1).Info("dataplane has no kuma.io/workload label, skipping")
		return kube_ctrl.Result{}, nil
	}

	meshName := dp.Mesh

	// Ensure Workload resource exists
	if err := r.createOrUpdateWorkload(ctx, workloadName, meshName, dp.Namespace); err != nil {
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
		// Set mesh label
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

func (r *WorkloadReconciler) cleanupOrphanedWorkloads(ctx context.Context, namespace string) error {
	log := r.Log.WithValues("namespace", namespace)

	// List all Workloads in the namespace
	workloads := &workload_k8s.WorkloadList{}
	if err := r.List(ctx, workloads, kube_client.InNamespace(namespace)); err != nil {
		return errors.Wrap(err, "unable to list Workloads")
	}

	// List all Dataplanes in the namespace
	dataplanes := &mesh_k8s.DataplaneList{}
	if err := r.List(ctx, dataplanes, kube_client.InNamespace(namespace)); err != nil {
		return errors.Wrap(err, "unable to list Dataplanes")
	}

	// Build a map of workload names that are still referenced
	referencedWorkloads := make(map[string]map[string]bool) // mesh -> workload name -> true
	for _, dp := range dataplanes.Items {
		if workloadName, ok := dp.GetLabels()[metadata.KumaWorkload]; ok && workloadName != "" {
			if referencedWorkloads[dp.Mesh] == nil {
				referencedWorkloads[dp.Mesh] = make(map[string]bool)
			}
			referencedWorkloads[dp.Mesh][workloadName] = true
		}
	}

	// Delete Workloads that are no longer referenced
	for _, workload := range workloads.Items {
		workloadName := workload.Name
		meshName := workload.GetMesh()

		if referencedWorkloads[meshName] == nil || !referencedWorkloads[meshName][workloadName] {
			log.Info("deleting orphaned workload", "workload", workloadName, "mesh", meshName)
			if err := r.Delete(ctx, &workload); err != nil && !kube_apierrs.IsNotFound(err) {
				return errors.Wrapf(err, "failed to delete orphaned Workload %s", workloadName)
			}
		}
	}

	return nil
}

func (r *WorkloadReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		Named("kuma-workload-controller").
		For(&mesh_k8s.Dataplane{}).
		Watches(&kube_core.Namespace{}, kube_handler.EnqueueRequestsFromMapFunc(NamespaceToDataplanesMapper(r.Log, mgr.GetClient())), builder.WithPredicates(predicate.LabelChangedPredicate{})).
		Watches(&mesh_k8s.Mesh{}, kube_handler.EnqueueRequestsFromMapFunc(MeshToAllDataplanesMapper(r.Log, mgr.GetClient())), builder.WithPredicates(CreateOrDeletePredicate{})).
		Complete(r)
}

func NamespaceToDataplanesMapper(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("namespace-to-dataplanes-mapper")
	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		dataplanes := &mesh_k8s.DataplaneList{}
		if err := client.List(ctx, dataplanes, kube_client.InNamespace(obj.GetName())); err != nil {
			l.WithValues("namespace", obj.GetName()).Error(err, "failed to fetch Dataplanes")
			return nil
		}
		var req []kube_reconcile.Request
		for _, dp := range dataplanes.Items {
			req = append(req, kube_reconcile.Request{
				NamespacedName: kube_types.NamespacedName{Namespace: dp.Namespace, Name: dp.Name},
			})
		}
		return req
	}
}

func MeshToAllDataplanesMapper(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("mesh-to-dataplanes-mapper")
	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		dataplanes := &mesh_k8s.DataplaneList{}
		if err := client.List(ctx, dataplanes); err != nil {
			l.WithValues("mesh", obj.GetName()).Error(err, "failed to fetch Dataplanes")
			return nil
		}
		var req []kube_reconcile.Request
		for _, dp := range dataplanes.Items {
			req = append(req, kube_reconcile.Request{
				NamespacedName: kube_types.NamespacedName{Namespace: dp.Namespace, Name: dp.Name},
			})
		}
		return req
	}
}
