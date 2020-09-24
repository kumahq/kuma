package controllers

import (
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
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

	dpProbes := &mesh_proto.Dataplane_Probes{
		Port: port,
	}
	for _, c := range pod.Spec.Containers {
		if c.LivenessProbe != nil && c.LivenessProbe.HTTPGet != nil {
			if endpoint, err := ProbeFor(c.LivenessProbe, port); err != nil {
				return nil, err
			} else {
				dpProbes.Endpoints = append(dpProbes.Endpoints, endpoint)
			}
		}
		if c.ReadinessProbe != nil && c.ReadinessProbe.HTTPGet != nil {
			if endpoint, err := ProbeFor(c.ReadinessProbe, port); err != nil {
				return nil, err
			} else {
				dpProbes.Endpoints = append(dpProbes.Endpoints, endpoint)
			}
		}
	}
	return dpProbes, nil
}

func ProbeFor(podProbe *kube_core.Probe, port uint32) (*mesh_proto.Dataplane_Probes_Endpoint, error) {
	inbound, err := probes.KumaProbe(*podProbe).ToReal(port)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert virtual probe to real")
	}
	return &mesh_proto.Dataplane_Probes_Endpoint{
		InboundPort: inbound.Port(),
		InboundPath: inbound.Path(),
		Path:        probes.KumaProbe(*podProbe).Path(),
	}, nil
}
