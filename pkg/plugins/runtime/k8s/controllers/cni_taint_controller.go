package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	nodeReadinessTaintKey = "NodeReadiness"
	nodeIndexField        = "spec.nodeName"
	cniAppLabel           = "app"
	cniPodNamespace       = "kube-system"
)

type CniNodeTaintReconciler struct {
	kube_client.Client
	Log logr.Logger

	CniApp string
}

func (r *CniNodeTaintReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	log := r.Log.WithValues("node", req.NamespacedName)

	// Fetch the Node instance
	node := &kube_core.Node{}
	if err := r.Get(ctx, req.NamespacedName, node); err != nil {
		if kube_apierrs.IsNotFound(err) {
			log.V(1).Info("node not found", "node", req.NamespacedName)
			return kube_ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch node")
		return kube_ctrl.Result{}, err
	}
	log.V(1).Info("node successfully fetched")

	kubeSystemPods := &kube_core.PodList{}
	namespaceOption := kube_client.InNamespace(cniPodNamespace)
	matchingFields := kube_client.MatchingFields{nodeIndexField: node.Name}
	matchingLabels := kube_client.MatchingLabels{cniAppLabel: r.CniApp}
	if err := r.Client.List(ctx, kubeSystemPods, namespaceOption, matchingFields, matchingLabels); err != nil {
		return kube_ctrl.Result{}, err
	}

	err := r.updateTaints(ctx, log, node, kubeSystemPods.Items)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update node taints")
	}

	return kube_ctrl.Result{}, err
}

func (r *CniNodeTaintReconciler) updateTaints(ctx context.Context, log logr.Logger, node *kube_core.Node, pods []kube_core.Pod) error {
	taintIndex := slices.IndexFunc(node.Spec.Taints, func(taint kube_core.Taint) bool {
		return taint.Key == nodeReadinessTaintKey && taint.Effect == kube_core.TaintEffectNoSchedule
	})

	if taintIndex >= 0 {
		if r.hasCniPodRunning(log, pods) {
			log.Info("has cni pod running and taint")
			return r.untaintNode(ctx, log, node, taintIndex)
		} else {
			log.Info("has no cni pod running and taint")
			return nil
		}
	} else {
		if r.hasCniPodRunning(log, pods) {
			log.V(1).Info("has cni pod running and no taint")
			return nil
		} else {
			log.Info("has no cni pod running and no taint")
			return r.taintNode(ctx, log, node)
		}
	}
}

func (r *CniNodeTaintReconciler) untaintNode(ctx context.Context, log logr.Logger, node *kube_core.Node, taintIndex int) error {
	node.Spec.Taints = slices.Delete(node.Spec.Taints, taintIndex, taintIndex+1)

	err := r.Client.Update(ctx, node)
	if err == nil {
		log.Info("removed the taint from node")
	}
	return err
}

func (r *CniNodeTaintReconciler) taintNode(ctx context.Context, log logr.Logger, node *kube_core.Node) error {
	node.Spec.Taints = append(node.Spec.Taints, kube_core.Taint{
		Key:    nodeReadinessTaintKey,
		Effect: kube_core.TaintEffectNoSchedule,
	})

	err := r.Client.Update(ctx, node)
	if err == nil {
		log.Info("added taint to node")
	}
	return err
}

func (r *CniNodeTaintReconciler) hasCniPodRunning(log logr.Logger, pods []kube_core.Pod) bool {
	for _, pod := range pods {
		isCniPod := pod.Labels[cniAppLabel] == r.CniApp
		if isCniPod && pod.Status.Phase == kube_core.PodRunning && isConditionTrue(pod.Status.Conditions, kube_core.PodReady) {
			log.V(1).Info("pod has kuma-cni-node running and ready", "pod", pod.Name, "status", pod.Status)
			return true
		}
	}
	return false
}

func (r *CniNodeTaintReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &kube_core.Pod{}, nodeIndexField, func(obj kube_client.Object) []string {
		pod := obj.(*kube_core.Pod)
		return []string{pod.Spec.NodeName}
	}); err != nil {
		return err
	}

	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.Node{}, builder.WithPredicates(nodeEvents)).
		Watches(
			&kube_source.Kind{Type: &kube_core.Pod{}},
			kube_handler.EnqueueRequestsFromMapFunc(podToNodeMapper(r.Log)),
			builder.WithPredicates(podEvents(r.CniApp)),
		).
		Complete(r)
}

func podToNodeMapper(log logr.Logger) kube_handler.MapFunc {
	return func(obj kube_client.Object) []kube_reconcile.Request {
		pod, ok := obj.(*kube_core.Pod)
		if !ok {
			log.WithValues("pod", obj.GetName()).Error(errors.Errorf("wrong argument type: expected %T, got %T", pod, obj), "wrong argument type")
			return nil
		}

		// we only care about new object but pod update event triggers two reconciliation (for old and new object)
		// and on the first one updateEvent.OldObject does not have a node assigned so that's why I'm filtering out here
		if pod.Spec.NodeName == "" {
			return nil
		}

		req := kube_reconcile.Request{NamespacedName: kube_types.NamespacedName{
			Name: pod.Spec.NodeName,
		}}
		return []kube_reconcile.Request{req}
	}
}

var nodeEvents = predicate.Funcs{
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

func podEvents(cniApp string) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return filterPods(deleteEvent.Object, cniApp)
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return filterPods(updateEvent.ObjectNew, cniApp)
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return false
		},
	}
}

func filterPods(obj kube_client.Object, cniApp string) bool {
	pod, ok := obj.(*kube_core.Pod)
	if !ok {
		return false
	}
	if pod.Spec.NodeName == "" {
		return false
	}
	if pod.Namespace != cniPodNamespace {
		return false
	}
	if pod.Labels[cniAppLabel] != cniApp {
		return false
	}

	return true
}

func isConditionTrue(conditions []kube_core.PodCondition, conditionType kube_core.PodConditionType) bool {
	for _, condition := range conditions {
		if condition.Type == conditionType && condition.Status == kube_core.ConditionTrue {
			return true
		}
	}

	return false
}
