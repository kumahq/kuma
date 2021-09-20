package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_record "k8s.io/client-go/tools/record"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	util_k8s "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

// PodStatusReconciler tracks pods status changes and signals kuma-dp when it has to complete
type PodStatusReconciler struct {
	kube_client.Client
	kube_record.EventRecorder
	Scheme            *kube_runtime.Scheme
	ResourceManager   manager.ResourceManager
	Log               logr.Logger
	ResourceConverter k8s_common.Converter
	EnvoyAdminClient  admin.EnvoyAdminClient
}

func (r *PodStatusReconciler) Reconcile(req kube_ctrl.Request) (kube_ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pod-status", req.NamespacedName)

	// Fetch the Pod instance
	pod := &kube_core.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Pod")
		return kube_ctrl.Result{}, err
	}

	// we process only Pods owned by a Job
	isJob := false
	for _, o := range pod.GetObjectMeta().GetOwnerReferences() {
		if o.Kind == "Job" {
			isJob = true
			break
		}
	}
	if !isJob {
		return kube_ctrl.Result{}, nil
	}

	hasSidecar := false
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == util_k8s.KumaSidecarContainerName {
			hasSidecar = true
			if cs.State.Terminated != nil {
				// the sidecar already terminated
				return kube_ctrl.Result{}, nil
			}
		} else if cs.State.Terminated == nil || cs.State.Terminated.ExitCode != 0 {
			// at least one non-sidecar container not terminated
			// or did not completed successfully
			// no need to tell envoy to quit yet
			return kube_ctrl.Result{}, nil
		}
	}

	if hasSidecar {
		log.V(1).Info("unterminated pod with sidecar found, sending quit to envoy")

		dataplane := &mesh_k8s.Dataplane{}
		if err := r.Get(ctx, req.NamespacedName, dataplane); err != nil {
			if kube_apierrs.IsNotFound(err) {
				log.V(1).Info("dataplane is not found", "name", req.NamespacedName)
				return kube_ctrl.Result{}, nil
			}
			log.Error(err, "unable to fetch Dataplane")
			return kube_ctrl.Result{}, err
		}

		dp := core_mesh.NewDataplaneResource()
		if err := r.ResourceConverter.ToCoreResource(dataplane, dp); err != nil {
			converterLog.Error(err, "failed to parse Dataplane", "dataplane", dataplane.Spec)
			return kube_ctrl.Result{}, err
		}

		var errs error
		err := r.EnvoyAdminClient.PostQuit(dp)
		if err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, "envoy admin client failed. Most probably the pod is already going down."))
		}

		err = r.Client.Delete(ctx, dataplane)
		if err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, "trying to delete the job's dataplane"))
		}

		return kube_ctrl.Result{}, errs
	}

	return kube_ctrl.Result{}, nil
}

func (r *PodStatusReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.Pod{}, builder.WithPredicates(podStatusEvents)).
		Complete(r)
}

// we only want status event updates
var podStatusEvents = predicate.Funcs{
	CreateFunc: func(event event.CreateEvent) bool {
		return false
	},
	DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
		return false
	},
	UpdateFunc: func(updateEvent event.UpdateEvent) bool {
		return true
	},
	GenericFunc: func(genericEvent event.GenericEvent) bool {
		return false
	},
}
