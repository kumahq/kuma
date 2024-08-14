package probes

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ProxiedApplicationProbe is a type which allows to manipulate Kubernetes HttpGet probes.
// Probe can be either Virtual or Real.
//
// Real probe is a probe provided by user. The only constraint existing for Real
// probes is that the port must be different from 'virtualPort'.
//
// Virtual probe is an automatically generated probe on the basis of the Real probe.
// If probe's port equal to 'virtualPort' and the first segment of probe's path is an integer
// then probe is a virtual probe.
type ProxiedApplicationProbe kube_core.Probe

func (p ProxiedApplicationProbe) ToVirtual(virtualPort uint32) (ProxiedApplicationProbe, error) {
	switch {
	case p.ProbeHandler.HTTPGet != nil:
		return p.httpProbeToVirtual(virtualPort)
	case p.ProbeHandler.TCPSocket != nil:
		return p.tcpProbeToVirtual(virtualPort)
	case p.ProbeHandler.GRPC != nil:
		return p.grpcProbeToVirtual(virtualPort)
	default:
		return ProxiedApplicationProbe{}, errors.New("unsupported probe type")
	}
}

func (p ProxiedApplicationProbe) httpProbeToVirtual(virtualPort uint32) (ProxiedApplicationProbe, error) {
	appPort := uint32(p.HTTPGet.Port.IntValue())
	if appPort == virtualPort {
		return ProxiedApplicationProbe{}, errors.Errorf("cannot override Pod's probes. Port for probe cannot "+
			"be set to %d. It is reserved for the dataplane that will serve pods without mTLS.", virtualPort)
	}

	probePath := p.Path()
	if !strings.HasPrefix(p.Path(), "/") {
		probePath = fmt.Sprintf("/%s", p.Path())
	}

	var headers []kube_core.HTTPHeader
	headerIdx := slices.IndexFunc(p.HTTPGet.HTTPHeaders, func(header kube_core.HTTPHeader) bool {
		return header.Name == "Host"
	})

	var hostHeader kube_core.HTTPHeader
	if headerIdx != -1 {
		hostHeader = p.HTTPGet.HTTPHeaders[headerIdx]
		headers = append(p.HTTPGet.HTTPHeaders[:headerIdx], HostHeader(hostHeader.Value))
		headers = append(headers, p.HTTPGet.HTTPHeaders[headerIdx+1:]...)
	} else {
		headers = append(headers, p.HTTPGet.HTTPHeaders...)
	}

	if p.HTTPGet.Scheme != "" && p.HTTPGet.Scheme != kube_core.URISchemeHTTP {
		headers = append(headers, SchemeHeader(p.HTTPGet.Scheme))
	}

	if p.TimeoutSeconds > 1 {
		headers = append(headers, TimeoutHeader(p.TimeoutSeconds))
	}

	return ProxiedApplicationProbe{
		ProbeHandler: kube_core.ProbeHandler{
			HTTPGet: &kube_core.HTTPGetAction{
				Port:        intstr.FromInt32(int32(virtualPort)),
				Path:        fmt.Sprintf("/%d%s", appPort, probePath),
				HTTPHeaders: headers,
			},
		},
	}, nil
}

func (p ProxiedApplicationProbe) tcpProbeToVirtual(virtualPort uint32) (ProxiedApplicationProbe, error) {
	appPort := uint32(p.TCPSocket.Port.IntValue())
	if appPort == virtualPort {
		return ProxiedApplicationProbe{}, errors.Errorf("cannot override Pod's probes. Port for probe cannot "+
			"be set to %d. It is reserved for the dataplane that will serve pods without mTLS.", virtualPort)
	}

	var headers []kube_core.HTTPHeader

	if p.TimeoutSeconds > 1 {
		headers = append(headers, TimeoutHeader(p.TimeoutSeconds))
	}

	return ProxiedApplicationProbe{
		ProbeHandler: kube_core.ProbeHandler{
			HTTPGet: &kube_core.HTTPGetAction{
				Port:        intstr.FromInt32(int32(virtualPort)),
				Path:        fmt.Sprintf("/tcp/%d", appPort),
				HTTPHeaders: headers,
			},
		},
	}, nil
}

func (p ProxiedApplicationProbe) grpcProbeToVirtual(virtualPort uint32) (ProxiedApplicationProbe, error) {
	appPort := uint32(p.GRPC.Port)
	if appPort == virtualPort {
		return ProxiedApplicationProbe{}, errors.Errorf("cannot override Pod's probes. Port for probe cannot "+
			"be set to %d. It is reserved for the dataplane that will serve pods without mTLS.", virtualPort)
	}

	var headers []kube_core.HTTPHeader

	if p.TimeoutSeconds > 1 {
		headers = append(headers, TimeoutHeader(p.TimeoutSeconds))
	}

	if p.GRPC.Service != nil && *p.GRPC.Service != "" {
		headers = append(headers, GRPCServiceHeader(*p.GRPC.Service))
	}

	return ProxiedApplicationProbe{
		ProbeHandler: kube_core.ProbeHandler{
			HTTPGet: &kube_core.HTTPGetAction{
				Port:        intstr.FromInt32(int32(virtualPort)),
				Path:        fmt.Sprintf("/grpc/%d", appPort),
				HTTPHeaders: headers,
			},
		},
	}, nil
}

func (p ProxiedApplicationProbe) Port() uint32 {
	switch {
	case p.ProbeHandler.HTTPGet != nil:
		return uint32(p.HTTPGet.Port.IntValue())
	case p.ProbeHandler.TCPSocket != nil:
		return uint32(p.TCPSocket.Port.IntValue())
	case p.ProbeHandler.GRPC != nil:
		return uint32(p.GRPC.Port)
	default:
		return 0
	}
}

func (p ProxiedApplicationProbe) Path() string {
	return p.HTTPGet.Path
}

func (p ProxiedApplicationProbe) Headers() []kube_core.HTTPHeader {
	return p.HTTPGet.HTTPHeaders
}

func (p ProxiedApplicationProbe) OverridingSupported() bool {
	switch {
	case p.ProbeHandler.HTTPGet != nil:
		return true
	case p.ProbeHandler.TCPSocket != nil:
		return true
	case p.ProbeHandler.GRPC != nil:
		return true
	default:
		return false
	}
}
