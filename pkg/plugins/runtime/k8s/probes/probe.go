package probes

import (
	"fmt"
	"strconv"
	"strings"

	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ProbePort = 9000
)

type KumaProbe kube_core.Probe

func (p KumaProbe) ToInbound() (KumaProbe, bool) {
	if p.Port() != ProbePort {
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

func (p KumaProbe) ToVirtual() KumaProbe {
	return KumaProbe{
		Handler: kube_core.Handler{
			HTTPGet: &kube_core.HTTPGetAction{
				Port: intstr.FromInt(ProbePort),
				Path: fmt.Sprintf("/%d%s", p.Port(), p.Path()),
			},
		},
	}
}

func (p KumaProbe) Port() int {
	return p.HTTPGet.Port.IntValue()
}

func (p KumaProbe) Path() string {
	return p.HTTPGet.Path
}
