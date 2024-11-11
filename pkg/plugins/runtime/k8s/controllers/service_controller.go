package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	kube_client.Client
	Log logr.Logger
}

// Reconcile is in charge of injecting "ingress.kubernetes.io/service-upstream" annotation to the Services
// that are in Kuma enabled namespaces
func (r *ServiceReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	log := r.Log.WithValues("service", req.NamespacedName)

	svc := &kube_core.Service{}
	if err := r.Get(ctx, req.NamespacedName, svc); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, errors.Wrapf(err, "unable to fetch Service %s", req.NamespacedName.Name)
	}

	if svc.GetAnnotations()[metadata.KumaGatewayAnnotation] == metadata.AnnotationBuiltin {
		return kube_ctrl.Result{}, nil
	}

	if svcPartOfMesh, err := r.isServicePartOfMesh(ctx, log, svc, req); err != nil {
		return kube_ctrl.Result{}, err
	} else if !svcPartOfMesh {
		log.V(1).Info(req.NamespacedName.String() + "is not part of the mesh")
		return kube_ctrl.Result{}, nil
	}

	if svc.Spec.Type == kube_core.ServiceTypeExternalName {
		log.V(1).Info(
			"ignoring Service because it is of an unsupported type",
			"name", req.NamespacedName.String(),
			"type", kube_core.ServiceTypeExternalName,
		)
		return kube_ctrl.Result{}, nil
	}

	log.Info("annotating service which is part of the mesh", "annotation", fmt.Sprintf("%s=%s", metadata.IngressServiceUpstream, metadata.AnnotationTrue))
	annotations := metadata.Annotations(svc.Annotations)
	if annotations == nil {
		annotations = metadata.Annotations{}
	}
	ignored, _, err := annotations.GetEnabled(metadata.KumaIgnoreAnnotation)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrapf(err, "unable to retrieve %s annotation for %s", metadata.KumaIgnoreAnnotation, svc.Name)
	}
	if ignored {
		return kube_ctrl.Result{}, nil
	}
	annotations[metadata.IngressServiceUpstream] = metadata.AnnotationTrue
	annotations[metadata.NginxIngressServiceUpstream] = metadata.AnnotationTrue
	svc.Annotations = annotations

	if err = r.Update(ctx, svc); err != nil {
		return kube_ctrl.Result{}, errors.Wrapf(err, "unable to update ingress service upstream annotation on service %s", svc.Name)
	}

	return kube_ctrl.Result{}, nil
}

func (r *ServiceReconciler) isServicePartOfMesh(ctx context.Context, log logr.Logger, svc *kube_core.Service, req kube_ctrl.Request) (bool, error) {
	namespace := &kube_core.Namespace{}
	if err := r.Get(ctx, kube_types.NamespacedName{Name: svc.GetNamespace()}, namespace); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "unable to fetch Namespace %s", req.NamespacedName.Name)
	}

	injectEnabledOnNs, _, err := metadata.Annotations(namespace.Labels).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
	if err != nil {
		return false, errors.Wrapf(err, "unable to check sidecar injection label on namespace %s", namespace.Name)
	}

	if !injectEnabledOnNs && svc.Spec.Selector != nil {
		pods := &kube_core.PodList{}
		if err := r.List(ctx, pods, kube_client.InNamespace(svc.GetNamespace()),
			kube_client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(svc.Spec.Selector)}); err != nil {
			return false, errors.Wrapf(err, "unable to list pods of Service %s", req.NamespacedName.Name)
		}

		injectEnabledOnPod := false
		for i := range pods.Items {
			pod := pods.Items[i]
			injectedEnabled, _, errLabel := metadata.Annotations(pod.Labels).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
			if errLabel == nil && pod.DeletionTimestamp == nil && injectedEnabled {
				injectEnabledOnPod = true
				break
			}
		}

		if !injectEnabledOnPod {
			log.V(1).Info(req.NamespacedName.String() + "is not part of the mesh")
			return false, nil
		}

		return true, nil
	}

	return false, nil
}

func (r *ServiceReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		Named("kuma-service-controller").
		For(&kube_core.Service{}, builder.WithPredicates(serviceEvents)).
		Complete(r)
}

// we only want create and update events
var serviceEvents = predicate.Funcs{
	CreateFunc: func(event event.CreateEvent) bool {
		return true
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
