package probes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// KumaProbe is a type which allows to manipulate Kubernetes HttpGet probes.
// Probe can be either Virtual or Real.
//
// Real probe is a probe provided by user. The only constraint existing for Real
// probes is that the port must be different from 'virtualPort'.
//
// Virtual probe is an automatically generated probe on the basis of the Real probe.
// If probe's port equal to 'virtualPort' and the first segment of probe's path is an integer
// then probe is a virtual probe.
type KumaProbe kube_core.Probe

// ToReal creates Real probe assuming that 'p' is a Virtual probe. If 'p' is a Real probe,
// then method returns an error
func (p KumaProbe) ToReal(virtualPort uint32) (KumaProbe, error) {
	if p.Port() != virtualPort {
		return KumaProbe{}, errors.Errorf("probe's port %d should be equal to virtual port %d", p.Port(), virtualPort)
	}
	segments := strings.Split(p.Path(), "/")
	if len(segments) < 2 || segments[0] != "" {
		return KumaProbe{}, errors.New("not enough segments in probe's path")
	}
	vport, err := strconv.ParseInt(segments[1], 10, 32)
	if err != nil {
		return KumaProbe{}, errors.Errorf("invalid port value %s", segments[1])
	}
	return KumaProbe{
		ProbeHandler: kube_core.ProbeHandler{
			HTTPGet: &kube_core.HTTPGetAction{
				Port: intstr.FromInt(int(vport)),
				Path: fmt.Sprintf("/%s", strings.Join(segments[2:], "/")),
			},
		},
	}, nil
}

func (p KumaProbe) ToVirtual(virtualPort uint32) (KumaProbe, error) {
	if p.Port() == virtualPort {
		return KumaProbe{}, errors.Errorf("cannot override Pod's probes. Port for probe cannot "+
			"be set to %d. It is reserved for the dataplane that will serve pods without mTLS.", virtualPort)
	}
	probePath := p.Path()
	if !strings.HasPrefix(p.Path(), "/") {
		probePath = fmt.Sprintf("/%s", p.Path())
	}
	return KumaProbe{
		ProbeHandler: kube_core.ProbeHandler{
			HTTPGet: &kube_core.HTTPGetAction{
				Port: intstr.FromInt(int(virtualPort)),
				Path: fmt.Sprintf("/%d%s", p.Port(), probePath),
			},
		},
	}, nil
}

func (p KumaProbe) Port() uint32 {
	return uint32(p.HTTPGet.Port.IntValue())
}

func (p KumaProbe) Path() string {
	return p.HTTPGet.Path
}
