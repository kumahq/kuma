package injector

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_types "k8s.io/apimachinery/pkg/types"

	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

func (i *KumaInjector) preCheck(ctx context.Context, pod *kube_core.Pod, logger logr.Logger) (string, error) {
	ns, err := i.namespaceFor(ctx, pod)
	if err != nil {
		return "", errors.Wrap(err, "could not retrieve namespace for pod")
	}

	// Log deprecated annotations
	for _, d := range metadata.PodAnnotationDeprecations {
		if _, exists := pod.Annotations[d.Key]; exists {
			logger.Info("WARNING: using deprecated pod annotation", "key", d.Key, "message", d.Message)
		}
	}
	logYesNoDeprecations(pod.Annotations, logger)

	if inject, err := i.needToInject(pod, ns); err != nil {
		return "", err
	} else if !inject {
		logger.V(1).Info("skipping Kuma injection")
		return "", nil
	}

	meshName := k8s_util.MeshOfByLabelOrAnnotation(logger, pod, ns)
	logger = logger.WithValues("mesh", meshName)
	// Check mesh exists
	if err := i.client.Get(ctx, kube_types.NamespacedName{Name: meshName}, &mesh_k8s.Mesh{}); err != nil {
		return "", err
	}

	// Warn if an init container in the pod is using the same UID as the sidecar. This traffic will be exempt from
	// redirection and may be unintended behavior.
	for _, c := range pod.Spec.InitContainers {
		if c.SecurityContext != nil && c.SecurityContext.RunAsUser != nil {
			if *c.SecurityContext.RunAsUser == i.cfg.SidecarContainer.UID {
				logger.Info(
					"WARNING: init container using ignored sidecar UID",
					"container",
					c.Name,
					"uid",
					i.cfg.SidecarContainer.UID,
				)
			}
		}
	}

	var duplicateUidContainers []string
	// Error if a container in the pod is using the same UID as the sidecar. This scenario is not supported.
	for _, c := range pod.Spec.Containers {
		if c.SecurityContext != nil && c.SecurityContext.RunAsUser != nil {
			if *c.SecurityContext.RunAsUser == i.cfg.SidecarContainer.UID {
				duplicateUidContainers = append(duplicateUidContainers, c.Name)
			}
		}
	}

	if len(duplicateUidContainers) > 0 {
		err := fmt.Errorf(
			"containers using same UID as sidecar is unsupported: %q",
			duplicateUidContainers,
		)

		logger.Error(err, "injection failed")

		return "", err
	}
	return meshName, nil
}

func (i *KumaInjector) needToInject(pod *kube_core.Pod, ns *kube_core.Namespace) (bool, error) {
	log.WithValues("name", pod.Name, "namespace", pod.Namespace)
	if i.isInjectionException(pod) {
		log.V(1).Info("pod fulfills exception requirements")
		return false, nil
	}

	for _, container := range append(append([]kube_core.Container{}, pod.Spec.Containers...), pod.Spec.InitContainers...) {
		if container.Name == k8s_util.KumaSidecarContainerName {
			log.V(1).Info("pod already has Kuma sidecar")
			return false, nil
		}
	}

	enabled, exist, err := metadata.Annotations(pod.Labels).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
	if err != nil {
		return false, err
	}
	if exist {
		if !enabled {
			log.V(1).Info(`pod has "kuma.io/sidecar-injection: disabled" label`)
		}
		return enabled, nil
	}

	enabled, exist, err = metadata.Annotations(ns.Labels).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
	if err != nil {
		return false, err
	}
	if exist {
		if !enabled {
			log.V(1).Info(`namespace has "kuma.io/sidecar-injection: disabled" label`)
		}
		return enabled, nil
	}
	return false, err
}

func (i *KumaInjector) isInjectionException(pod *kube_core.Pod) bool {
	for key, value := range i.cfg.Exceptions.Labels {
		podValue, exist := pod.Labels[key]
		if exist && (value == "*" || value == podValue) {
			return true
		}
	}
	return false
}

// these are all existing annotations that are supported by metadata.GetBooleanWithDefault
var booleanAnnotations = map[string]bool{
	metadata.KumaTrafficDropInvalidPackets:         true,
	metadata.KumaTrafficIptablesLogs:               true,
	metadata.KumaWaitForDataplaneReady:             true,
	metadata.KumaTransparentProxyingEbpf:           true,
	metadata.KumaBuiltinDNS:                        true,
	metadata.KumaBuiltinDNSLogging:                 true,
	metadata.KumaGatewayAnnotation:                 true,
	metadata.KumaSidecarInjectionAnnotation:        true,
	metadata.KumaIngressAnnotation:                 true,
	metadata.KumaEgressAnnotation:                  true,
	metadata.KumaSidecarInjectedAnnotation:         true,
	metadata.KumaTransparentProxyingAnnotation:     true,
	metadata.KumaMetricsPrometheusAggregateEnabled: true,
	metadata.KumaVirtualProbesAnnotation:           true,
	metadata.KumaIgnoreAnnotation:                  true,
	metadata.KumaInitFirst:                         true,
}

func logYesNoDeprecations(podAnnotations map[string]string, logger logr.Logger) {
	for key, value := range podAnnotations {
		if _, isBooleanAnno := booleanAnnotations[key]; !isBooleanAnno {
			continue
		}

		if value == "yes" || value == "no" {
			replacement := "true"
			if value == "no" {
				replacement = "false"
			}
			logger.Info(fmt.Sprintf("WARNING: using '%s' for annotation '%s' is deprecated, please use '%s' instead",
				value, key, replacement))
		}
	}
}
