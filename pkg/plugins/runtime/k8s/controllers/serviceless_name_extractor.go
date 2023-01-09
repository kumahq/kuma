package controllers

import (
	"context"

	kube_apps "k8s.io/api/apps/v1"
	kube_batch "k8s.io/api/batch/v1"
	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
)

type NameExtractor struct {
	ReplicaSetGetter kube_client.Reader
	JobGetter        kube_client.Reader
}

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
				return rs.Name, rs.Kind, nil
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
				return cj.Name, cj.Kind, nil
			}
			jobOwners := cj.GetObjectMeta().GetOwnerReferences()
			for _, o := range jobOwners {
				if o.Kind == "CronJob" {
					return o.Name, o.Kind, nil
				}
			}
		default:
			return owner.Name, owner.Kind, nil
		}
	}
	return pod.Name, pod.Kind, nil
}
