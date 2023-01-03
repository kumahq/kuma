package controllers

import (
	"context"
	"strings"

	kube_apps "k8s.io/api/apps/v1"
	kube_batch "k8s.io/api/batch/v1"
	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
)

type NameExtractor struct {
	ReplicaSetGetter kube_client.Reader
	JobGetter        kube_client.Reader
}

// Getting the name of serviceless pods could be done by getting it owner.
// If we apply this pattern for all the cases it isn't going to be backward
// compatible. There are only 2 cases where names are consistent with old way
// Deployments and CronJob.
func (n *NameExtractor) Name(ctx context.Context, pod *kube_core.Pod) (string, string, error) {
	owners := pod.GetObjectMeta().GetOwnerReferences()
	namespace := pod.Namespace
	for _, owner := range owners {
		switch owner.Kind {
		case "ReplicaSet":
			rs := &kube_apps.ReplicaSet{}
			rsKey := kube_client.ObjectKey{Namespace: namespace, Name: owner.Name}
			if err := n.ReplicaSetGetter.Get(ctx, rsKey, rs); err != nil {
				return "", "", err
			}
			if len(rs.OwnerReferences) == 0 {
				// backwards compatibility
				return nameFromPod(pod), pod.Kind, nil
			}
			rsOwners := rs.GetObjectMeta().GetOwnerReferences()
			for _, o := range rsOwners {
				if o.Kind == "Deployment" {
					return o.Name, o.Kind, nil
				}
			}
		case "Job":
			cj := &kube_batch.Job{}
			cjKey := kube_client.ObjectKey{Namespace: namespace, Name: owner.Name}
			if err := n.JobGetter.Get(ctx, cjKey, cj); err != nil {
				return "", "", err
			}
			if len(cj.OwnerReferences) == 0 {
				// backwards compatibility
				return nameFromPod(pod), pod.Kind, nil
			}
			jobOwners := cj.GetObjectMeta().GetOwnerReferences()
			for _, o := range jobOwners {
				if o.Kind == "CronJob" {
					return o.Name, o.Kind, nil
				}
			}
		default:
			// backwards compatibility
			return nameFromPod(pod), pod.Kind, nil
		}
	}
	// backwards compatibility
	return nameFromPod(pod), pod.Kind, nil
}

func nameFromPod(pod *kube_core.Pod) string {
	// the name is in format <name>-<replica set id>-<pod id>
	// this is only valid if pod is managed by CronJob or Deployment
	split := strings.Split(pod.Name, "-")
	if len(split) > 2 {
		split = split[:len(split)-2]
	}

	return strings.Join(split, "-")
}
