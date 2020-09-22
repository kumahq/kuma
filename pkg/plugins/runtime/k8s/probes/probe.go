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
// then second return value is 'false'.
func (p KumaProbe) ToReal(virtualPort uint32) (KumaProbe, bool) {
	if p.Port() != virtualPort {
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

func (p KumaProbe) ToVirtual(virtualPort uint32) (KumaProbe, error) {
	if p.Port() == virtualPort {
		return KumaProbe{}, errors.Errorf("cannot override Pod's probes. Port for probe cannot "+
			"be set to %d. It is reserved for the dataplane that will serve pods without mTLS.", virtualPort)
	}
	return KumaProbe{
		Handler: kube_core.Handler{
			HTTPGet: &kube_core.HTTPGetAction{
				Port: intstr.FromInt(int(virtualPort)),
				Path: fmt.Sprintf("/%d%s", p.Port(), p.Path()),
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
