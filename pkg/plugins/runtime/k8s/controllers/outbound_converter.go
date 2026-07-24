package controllers

import (
	kube_core "k8s.io/api/core/v1"
)

func isHeadlessService(svc *kube_core.Service) bool {
	return svc.Spec.ClusterIP == kube_core.ClusterIPNone
}
