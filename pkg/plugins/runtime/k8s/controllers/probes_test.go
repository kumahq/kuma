package controllers

import (
	"fmt"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("ProbesFor", func() {

	It("should generate Dataplane Probes section", func() {
		pod := &kube_core.Pod{
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{
					metadata.KumaVirtualProbesAnnotation:     metadata.AnnotationEnabled,
					metadata.KumaVirtualProbesPortAnnotation: "10101",
				},
			},
			Spec: kube_core.PodSpec{
				Containers: []kube_core.Container{
					{
						LivenessProbe: &kube_core.Probe{Handler: kube_core.Handler{HTTPGet: &kube_core.HTTPGetAction{
							Port: intstr.FromInt(10101),
							Path: "/8080/live_status",
						}}},
						ReadinessProbe: &kube_core.Probe{Handler: kube_core.Handler{HTTPGet: &kube_core.HTTPGetAction{
							Port: intstr.FromInt(10101),
							Path: "/8081/ready_status",
						}}},
					},
				},
			},
		}
		inbounds := []*v1alpha1.Dataplane_Networking_Inbound{
			{Port: 8080},
			{Port: 8081},
		}
		dpProbes, err := ProbesFor(pod, inbounds)
		Expect(err).ToNot(HaveOccurred())

		fmt.Println(dpProbes)
	})
})
