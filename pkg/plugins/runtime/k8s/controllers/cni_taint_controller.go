package controllers

import (
	"context"
	"strings"

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

const nodeReadinessTaintKey = "NodeReadiness"

type NodeReconciler struct {
	kube_client.Client
	Log logr.Logger
}

func (r *NodeReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	log := r.Log.WithValues("node", req.NamespacedName)
	log.Info("event received")

	// Fetch the Node instance
	node := &kube_core.Node{}
	if err := r.Get(ctx, req.NamespacedName, node); err != nil {
		if kube_apierrs.IsNotFound(err) {
			log.Error(err, "node not found")
			return kube_ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch node")
		return kube_ctrl.Result{}, err
	}
	log.Info("node successfully fetched")

	// List Pods in on the node
	// can we use use: r.Client.Get(ctx, ) instead? I don't think we can it only allows for filtering namespacedName
	// matchingFields := kube_client.MatchingFields{
	//	"spec.nodeName": node.Name,
	//	"metadata.labels.parent-app": "kuma-cni",
	//}

	kubeSystemPods := &kube_core.PodList{}
	if err := r.Client.List(ctx, kubeSystemPods, kube_client.InNamespace("kube-system")); err != nil {
		return kube_ctrl.Result{}, err
	}
	var podsOnThisNode []kube_core.Pod
	for _, pod := range kubeSystemPods.Items {
		if pod.Spec.NodeName == node.Name {
			podsOnThisNode = append(podsOnThisNode, pod)
		}
	}

	err := r.updateTaints(ctx, log, node, podsOnThisNode)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update node taints")
	}

	return kube_ctrl.Result{}, err
}

func (r *NodeReconciler) updateTaints(ctx context.Context, log logr.Logger, node *kube_core.Node, pods []kube_core.Pod) error {
	if hasTaint(node) {
		if hasCniPodRunning(log, pods) {
			log.Info("has cni pod running and taint")
			return r.untaintNode(ctx, log, node)
		} else {
			log.Info("has no cni pod running and taint")
			return nil
		}
	} else {
		if hasCniPodRunning(log, pods) {
			log.Info("has cni pod running and no taint")
			return nil
		} else {
			log.Info("has no cni pod running and no taint")
			return r.taintNode(ctx, log, node)
		}
	}
}

func (r *NodeReconciler) untaintNode(ctx context.Context, log logr.Logger, node *kube_core.Node) error {
	taintIndex := slices.IndexFunc(node.Spec.Taints, func(taint kube_core.Taint) bool {
		return taint.Key == nodeReadinessTaintKey && taint.Effect == kube_core.TaintEffectNoSchedule
	})

	if taintIndex >= 0 {
		node.Spec.Taints = removeTaint(node.Spec.Taints, taintIndex)
	}

	err := r.Client.Update(ctx, node)
	if err == nil {
		log.Info("removed the taint from node")
	}
	return err
}

func removeTaint(s []kube_core.Taint, index int) []kube_core.Taint {
	return append(s[:index], s[index+1:]...)
}

func (r *NodeReconciler) taintNode(ctx context.Context, log logr.Logger, node *kube_core.Node) error {
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

func hasTaint(node *kube_core.Node) bool {
	foundTaint := false
	for _, taint := range node.Spec.Taints {
		if taint.Key == "NodeReadiness" && taint.Effect == kube_core.TaintEffectNoSchedule {
			foundTaint = true
		}
	}
	return foundTaint
}

func hasCniPodRunning(log logr.Logger, pods []kube_core.Pod) bool {
	podReady := false
	containersReady := false
	for _, pod := range pods {
		if strings.Contains(pod.Name, "kuma-cni-node") && pod.Status.Phase == kube_core.PodRunning {
			for _, condition := range pod.Status.Conditions {
				if condition.Type == kube_core.PodReady && condition.Status == kube_core.ConditionTrue {
					podReady = true
				}
				if condition.Type == kube_core.ContainersReady && condition.Status == kube_core.ConditionTrue {
					containersReady = true
				}
			}
			if podReady && containersReady {
				log.Info("pod has kuma-cni-node running and ready", "pod", pod.Name, "status", pod.Status)
				return true
			} else {
				podReady = false
				containersReady = false
			}
		}
	}
	return false
}

func (r *NodeReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.Node{}, builder.WithPredicates(nodeEvents)).
		// check this is necessary
		Watches(
			&kube_source.Kind{Type: &kube_core.Pod{}},
			kube_handler.EnqueueRequestsFromMapFunc(podToNodeMapper(r.Log)),
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
