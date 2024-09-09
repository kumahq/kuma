package controllers

import (
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

func ProbesFor(pod *kube_core.Pod) (*mesh_proto.Dataplane_Probes, error) {
	enabled, exist, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaVirtualProbesAnnotation)
	if err != nil {
		return nil, err
	}
	if !exist || !enabled {
		return nil, nil
	}

	port, exist, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaVirtualProbesPortAnnotation)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.Errorf("%s annotation doesn't exist", metadata.KumaVirtualProbesPortAnnotation)
	}

	probeProxyPort, _, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaApplicationProbeProxyPortAnnotation)
	if err != nil {
		return nil, err
	}

	dpProbes := &mesh_proto.Dataplane_Probes{
		Port: port,
	}
	var containersNeedingProbes []kube_core.Container

	var initContainerComesAfterKumaSidecar bool
	for _, c := range pod.Spec.InitContainers {
		if c.Name == util.KumaSidecarContainerName {
			initContainerComesAfterKumaSidecar = true
			continue
		}
		if initContainerComesAfterKumaSidecar && c.RestartPolicy != nil && *c.RestartPolicy == kube_core.ContainerRestartPolicyAlways {
			containersNeedingProbes = append(containersNeedingProbes, c)
		}
	}
	for _, c := range pod.Spec.Containers {
		if c.Name != util.KumaSidecarContainerName {
			containersNeedingProbes = append(containersNeedingProbes, c)
		}
	}

	for _, c := range containersNeedingProbes {
		if c.LivenessProbe != nil && c.LivenessProbe.HTTPGet != nil {
			if endpoint, err := probeFor(c.LivenessProbe, port, probeProxyPort); err != nil {
				return nil, err
			} else if endpoint != nil {
				dpProbes.Endpoints = append(dpProbes.Endpoints, endpoint)
			}
		}
		if c.ReadinessProbe != nil && c.ReadinessProbe.HTTPGet != nil {
			if endpoint, err := probeFor(c.ReadinessProbe, port, probeProxyPort); err != nil {
				return nil, err
			} else if endpoint != nil {
				dpProbes.Endpoints = append(dpProbes.Endpoints, endpoint)
			}
		}
		if c.StartupProbe != nil && c.StartupProbe.HTTPGet != nil {
			if endpoint, err := probeFor(c.StartupProbe, port, probeProxyPort); err != nil {
				return nil, err
			} else if endpoint != nil {
				dpProbes.Endpoints = append(dpProbes.Endpoints, endpoint)
			}
		}
	}

	return dpProbes, nil
}

func probeFor(podProbe *kube_core.Probe, port uint32, probeProxyPort uint32) (*mesh_proto.Dataplane_Probes_Endpoint, error) {
	// if is using application probe proxy, we override it back to the virtual probes port to pass the validation
	kumaProbe := probes.KumaProbe(*podProbe)
	if probeProxyPort > 0 && kumaProbe.Port() == probeProxyPort {
		kumaProbe.HTTPGet.Port = intstr.FromInt32(int32(port))
	}

	inbound, err := kumaProbe.ToReal(port)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert virtual probe to real")
	}

	if inbound.HTTPGet == nil {
		return nil, nil
	}

	return &mesh_proto.Dataplane_Probes_Endpoint{
		InboundPort: inbound.Port(),
		InboundPath: inbound.Path(),
		Path:        kumaProbe.Path(),
	}, nil
}
