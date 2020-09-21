package probes

import (
	"fmt"
	"strconv"
	"strings"

	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type KumaProbe kube_core.Probe

func (p KumaProbe) ToInbound(port uint32) (KumaProbe, bool) {
	if p.Port() != port {
		return KumaProbe{}, false
	}
	segments := strings.Split(p.Path(), "/")
	if len(segments) < 2 || segments[0] != "" {
		return KumaProbe{}, false
	}
	vport, err := strconv.ParseInt(segments[1], 10, 32)
	if err != nil {
		return KumaProbe{}, false
	}
	return KumaProbe{
		Handler: kube_core.Handler{
			HTTPGet: &kube_core.HTTPGetAction{
				Port: intstr.FromInt(int(vport)),
				Path: fmt.Sprintf("/%s", strings.Join(segments[2:], "/")),
			},
		},
	}, true
}

func (p KumaProbe) ToVirtual(port uint32) KumaProbe {
	return KumaProbe{
		Handler: kube_core.Handler{
			HTTPGet: &kube_core.HTTPGetAction{
				Port: intstr.FromInt(int(port)),
				Path: fmt.Sprintf("/%d%s", p.Port(), p.Path()),
			},
		},
	}
}

func (p KumaProbe) Port() uint32 {
	return uint32(p.HTTPGet.Port.IntValue())
}

func (p KumaProbe) Path() string {
	return p.HTTPGet.Path
}
