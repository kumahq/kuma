package k8s

import (
	chartcommon "helm.sh/helm/v4/pkg/chart/common"
	releaseutilv1 "helm.sh/helm/v4/pkg/release/v1/util"

	"github.com/kumahq/kuma/v2/pkg/util/data"
)

// customInstallOrder extends Helm v4's default InstallOrder with Gateway API resources.
// In Helm v4, MutatingWebhookConfiguration and ValidatingWebhookConfiguration were added
// to the InstallOrder, but GatewayClass was not included. This causes GatewayClass to be
// sorted alphabetically after webhook configurations, breaking Gateway functionality.
// GatewayClass must be installed before webhooks to ensure proper Gateway resource
// initialization and reconciliation.
var customInstallOrder = releaseutilv1.KindSortOrder{
	"PriorityClass",
	"Namespace",
	"NetworkPolicy",
	"ResourceQuota",
	"LimitRange",
	"PodSecurityPolicy",
	"PodDisruptionBudget",
	"ServiceAccount",
	"Secret",
	"SecretList",
	"ConfigMap",
	"StorageClass",
	"PersistentVolume",
	"PersistentVolumeClaim",
	"CustomResourceDefinition",
	"ClusterRole",
	"ClusterRoleList",
	"ClusterRoleBinding",
	"ClusterRoleBindingList",
	"Role",
	"RoleList",
	"RoleBinding",
	"RoleBindingList",
	"Service",
	"DaemonSet",
	"Pod",
	"ReplicationController",
	"ReplicaSet",
	"Deployment",
	"HorizontalPodAutoscaler",
	"StatefulSet",
	"Job",
	"CronJob",
	"IngressClass",
	"Ingress",
	"APIService",
	"GatewayClass",
	"MutatingWebhookConfiguration",
	"ValidatingWebhookConfiguration",
}

func SortResourcesByKind(files []data.File, kindsToSkip ...string) ([]data.File, error) {
	skippedKinds := map[string]struct{}{}
	for _, k := range kindsToSkip {
		skippedKinds[k] = struct{}{}
	}
	singleFile := data.JoinYAML(files)
	resources := releaseutilv1.SplitManifests(string(singleFile.Data))

	hooks, manifests, err := releaseutilv1.SortManifests(resources, chartcommon.VersionSet{"v1", "v1beta1", "v1alpha1"}, customInstallOrder)
	if err != nil {
		return nil, err
	}

	result := make([]data.File, 0, len(manifests)+len(hooks))
	for _, manifest := range manifests {
		if _, ok := skippedKinds[manifest.Head.Kind]; !ok {
			result = append(result, data.File{Data: []byte(manifest.Content)})
		}
	}

	for _, hook := range hooks {
		result = append(result, data.File{Data: []byte(hook.Manifest)})
	}
	return result, nil
}
