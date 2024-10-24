package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
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
// but only when Kuma isn't using the SidecarContainer feature
type PodStatusReconciler struct {
	kube_client.Client
	kube_record.EventRecorder
	Scheme            *kube_runtime.Scheme
	ResourceManager   manager.ResourceManager
	Log               logr.Logger
	ResourceConverter k8s_common.Converter
	EnvoyAdminClient  admin.EnvoyAdminClient
}

func (r *PodStatusReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	log := r.Log.WithValues("pod", req.NamespacedName)

	// Fetch the Pod instance
	pod := &kube_core.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Pod")
		return kube_ctrl.Result{}, err
	}

	dataplane := &mesh_k8s.Dataplane{}
	if err := r.Get(ctx, req.NamespacedName, dataplane); err != nil {
		if kube_apierrs.IsNotFound(err) {
			log.V(1).Info("dataplane not found")
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

	log.Info("sending request to terminate Envoy")
	if err := r.EnvoyAdminClient.PostQuit(ctx, dp); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "envoy admin client failed. Most probably the pod is already going down.")
	}
	return kube_ctrl.Result{}, nil
}

func (r *PodStatusReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		Named("kuma-pod-status-controller").
		For(&kube_core.Pod{}, builder.WithPredicates(
			onlyUpdates,
			onlySidecarContainerRunning,
		)).
		Complete(r)
}

var onlyUpdates = predicate.Funcs{
	CreateFunc: func(event event.CreateEvent) bool {
		return true // we need it in case of CP restart
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

var onlySidecarContainerRunning = predicate.NewPredicateFuncs(
	func(obj kube_client.Object) bool {
		pod := obj.(*kube_core.Pod)
		sidecarContainerRunning := false
		if pod.Spec.RestartPolicy == kube_core.RestartPolicyAlways {
			return false
		}

		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == util_k8s.KumaSidecarContainerName {
				if cs.State.Terminated != nil {
					return false
				}
				sidecarContainerRunning = true
			} else {
				switch pod.Spec.RestartPolicy {
				case kube_core.RestartPolicyNever:
					if cs.State.Terminated == nil {
						// at least one non-sidecar container not terminated
						// no need to tell envoy to quit yet
						return false
					}
				default:
					if cs.State.Terminated == nil || cs.State.Terminated.ExitCode != 0 {
						// at least one non-sidecar container not terminated
						// or did not completed successfully
						// no need to tell envoy to quit yet
						return false
					}
				}
			}
		}

		return sidecarContainerRunning
	},
)
