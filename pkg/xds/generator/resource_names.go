package generator

import (
	"fmt"
	"sort"
	"strings"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
)

func localClusterName(port uint32) string {
	return fmt.Sprintf("localhost:%d", port)
}

func inboundListenerName(address string, port uint32) string {
	return fmt.Sprintf("inbound:%s:%d", address, port)
}

func outboundListenerName(address string, port uint32) string {
	return fmt.Sprintf("outbound:%s:%d", address, port)
}

func outboundRouteName(service string) string {
	return fmt.Sprintf("outbound:%s", service)
}

func envoyAdminClusterName() string {
	return "kuma:envoy:admin"
}

func prometheusListenerName() string {
	return "kuma:metrics:prometheus"
}

func destinationClusterName(service string, selector map[string]string) string {
	var pairs []string
	for key, value := range selector {
		if key == mesh_proto.ServiceTag {
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}
	if len(pairs) == 0 {
		return service
	}
	sort.Strings(pairs)
	return fmt.Sprintf("%s{%s}", service, strings.Join(pairs, ","))
}
