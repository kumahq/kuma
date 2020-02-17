package names

import (
	"fmt"
	"sort"
	"strings"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
)

func GetLocalClusterName(port uint32) string {
	return fmt.Sprintf("localhost:%d", port)
}

func GetInboundListenerName(address string, port uint32) string {
	return fmt.Sprintf("inbound:%s:%d", address, port)
}

func GetOutboundListenerName(address string, port uint32) string {
	return fmt.Sprintf("outbound:%s:%d", address, port)
}

func GetInboundRouteName(service string) string {
	return fmt.Sprintf("inbound:%s", service)
}

func GetOutboundRouteName(service string) string {
	return fmt.Sprintf("outbound:%s", service)
}

func GetEnvoyAdminClusterName() string {
	return "kuma:envoy:admin"
}

func GetPrometheusListenerName() string {
	return "kuma:metrics:prometheus"
}

func GetDestinationClusterName(service string, selector map[string]string) string {
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
