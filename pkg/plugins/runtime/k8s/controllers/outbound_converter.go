package controllers

import (
	"context"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
)

func isHeadlessService(svc *kube_core.Service) bool {
	return svc.Spec.ClusterIP == kube_core.ClusterIPNone
}

// Services of ExternalName type should not have any selectors.
// Kubernetes does not validate this, so in rare cases, a service of
// ExternalName type could point to a workload inside the mesh. If this
// happens, we will add the service to the VIPs config map, but we will
// not be able to obtain its IP address. As a result, the key in the map
// will be incorrect (e.g., "1:"). We do not currently support
// ExternalName services, so we can safely skip them from processing.
func isExternalNameService(svc *kube_core.Service) bool {
	return svc != nil && svc.Spec.Type == kube_core.ServiceTypeExternalName
}

func isServiceLess(port uint32) bool {
	return port == mesh_proto.TCPPortReserved
}

func k8sService(ctx context.Context, serviceTag string, client kube_client.Reader) (*kube_core.Service, uint32, error) {
	name, ns, port, err := parseService(serviceTag)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "failed to parse `service` host %q as FQDN", serviceTag)
	}
	if isServiceLess(port) {
		return nil, port, nil
	}

	svc := &kube_core.Service{}
	svcKey := kube_client.ObjectKey{Namespace: ns, Name: name}
	if err := client.Get(ctx, svcKey, svc); err != nil {
		return nil, 0, errors.Wrapf(err, "failed to get Service %q", svcKey)
	}
	return svc, port, nil
}

func parseService(host string) (string, string, uint32, error) {
	// split host into <name>_<namespace>_svc_<port>
	segments := strings.Split(host, "_")

	var port uint32
	switch len(segments) {
	case 4:
		p, err := strconv.ParseInt(segments[3], 10, 32)
		if err != nil {
			return "", "", 0, err
		}
		port = uint32(p)
	case 3:
		// service less service names have no port, so we just put the reserved
		// one here to note that this service is actually
		port = mesh_proto.TCPPortReserved
	default:
		return "", "", 0, errors.Errorf("service tag in unexpected format")
	}

	name, namespace := segments[0], segments[1]
	return name, namespace, port, nil
}
