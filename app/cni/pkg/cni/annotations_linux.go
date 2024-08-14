package cni

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	defaultProxyStatusPort     = "9901"
	defaultOutboundPort        = "15001"
	defaultInboundPort         = "15006"
	defaultIPFamilyMode        = "dualstack"
	defaultBuiltinDNSPort      = "15053"
	defaultNoRedirectUID       = "5678"
	defaultAppProbeProxyPort   = "9000"
	defaultRedirectExcludePort = defaultProxyStatusPort
)

var annotationRegistry = map[string]*annotationParam{
	"inject":                      {"kuma.io/sidecar-injection", "", alwaysValidFunc},
	"ports":                       {"kuma.io/envoy-admin-port", "", validatePortList},
	"excludeInboundPorts":         {"traffic.kuma.io/exclude-inbound-ports", defaultRedirectExcludePort, validatePortList},
	"excludeOutboundPorts":        {"traffic.kuma.io/exclude-outbound-ports", defaultRedirectExcludePort, validatePortList},
	"inboundPort":                 {"kuma.io/transparent-proxying-inbound-port", defaultInboundPort, validatePortList},
	"ipFamilyMode":                {"kuma.io/transparent-proxying-ip-family-mode", defaultIPFamilyMode, validateIpFamilyMode},
	"outboundPort":                {"kuma.io/transparent-proxying-outbound-port", defaultOutboundPort, validatePortList},
	"isGateway":                   {"kuma.io/gateway", "false", alwaysValidFunc},
	"builtinDNS":                  {"kuma.io/builtin-dns", "false", alwaysValidFunc},
	"builtinDNSPort":              {"kuma.io/builtin-dns-port", defaultBuiltinDNSPort, validatePortList},
	"excludeOutboundPortsForUIDs": {"traffic.kuma.io/exclude-outbound-ports-for-uids", "", alwaysValidFunc},
	"noRedirectUID":               {"kuma.io/sidecar-uid", defaultNoRedirectUID, alwaysValidFunc},
	"dropInvalidPackets":          {"traffic.kuma.io/drop-invalid-packets", "false", alwaysValidFunc},
	"iptablesLogs":                {"traffic.kuma.io/iptables-logs", "false", alwaysValidFunc},
	"excludeInboundIPs":           {"traffic.kuma.io/exclude-inbound-ips", "", validateIPs},
	"excludeOutboundIPs":          {"traffic.kuma.io/exclude-outbound-ips", "", validateIPs},
	"applicationProbeProxyPort":   {"kuma.io/application-probe-proxy-port", defaultAppProbeProxyPort, validateSinglePort},
}

type IntermediateConfig struct {
	// while https://github.com/kumahq/kuma/issues/8324 is not implemented, when changing the config,
	// keep in mind to update all other places listed in the issue

	targetPort                  string
	inboundPort                 string
	ipFamilyMode                string
	noRedirectUID               string
	excludeInboundPorts         string
	excludeOutboundPorts        string
	excludeOutboundPortsForUIDs string
	isGateway                   string
	builtinDNS                  string
	builtinDNSPort              string
	dropInvalidPackets          string
	iptablesLogs                string
	excludeInboundIPs           string
	excludeOutboundIPs          string
}

type annotationValidationFunc func(value string) error

type annotationParam struct {
	key        string
	defaultVal string
	validator  annotationValidationFunc
}

func alwaysValidFunc(_ string) error {
	return nil
}

func splitPorts(portsString string) []string {
	return strings.Split(portsString, ",")
}

func parsePort(portStr string) (uint16, error) {
	port, err := strconv.ParseUint(strings.TrimSpace(portStr), 10, 16)
	if err != nil {
		return 0, errors.Wrapf(err, "failed parsing port %q", portStr)
	}
	return uint16(port), nil
}

func parsePorts(portsString string) ([]int, error) {
	portsString = strings.TrimSpace(portsString)
	ports := make([]int, 0)
	if len(portsString) > 0 {
		for _, portStr := range splitPorts(portsString) {
			port, err := parsePort(portStr)
			if err != nil {
				return nil, err
			}
			ports = append(ports, int(port))
		}
	}
	return ports, nil
}

// validateIPs checks if the input string contains valid IP addresses or CIDR
// notations. It accepts a comma-separated string of IP addresses and/or CIDR
// blocks, trims any surrounding whitespace, and validates each entry.
//
// Args:
//   - addresses (string): A comma-separated string of IP addresses or CIDR
//     blocks.
//
// Returns:
//   - error: An error if the input string is empty or if any of the IP
//     addresses or CIDR blocks are invalid.
func validateIPs(addresses string) error {
	addresses = strings.TrimSpace(addresses)

	if addresses == "" {
		return errors.New("IPs cannot be empty")
	}

	// Split the string into individual addresses based on commas.
	for _, address := range strings.Split(addresses, ",") {
		address = strings.TrimSpace(address)

		// Check if the address is a valid CIDR block.
		if _, _, err := net.ParseCIDR(address); err == nil {
			continue
		}

		// Check if the address is a valid IP address.
		if ip := net.ParseIP(address); ip != nil {
			continue
		}

		// If the address is neither a valid IP nor a valid CIDR block, return
		// an error.
		return errors.Errorf(
			"invalid IP address: '%s'. Expected format: <ip> or <ip>/<cidr> "+
				"(e.g., 10.0.0.1, 172.16.0.0/16, fe80::1, fe80::/10)",
			address,
		)
	}

	return nil
}

func validatePortList(ports string) error {
	if _, err := parsePorts(ports); err != nil {
		return errors.Wrapf(err, "portList %q", ports)
	}
	return nil
}

func validateSinglePort(portString string) error {
	if _, err := parsePort(portString); err != nil {
		return err
	}
	return nil
}

func validateIpFamilyMode(val string) error {
	if val == "" {
		return errors.New("value is empty")
	}

	validValues := []string{"dualstack", "ipv4", "ipv6"}
	for _, valid := range validValues {
		if valid == val {
			return nil
		}
	}
	return errors.New(fmt.Sprintf("value '%s' is not a valid IP family mode", val))
}

func getAnnotationOrDefault(name string, annotations map[string]string) (string, error) {
	if _, ok := annotationRegistry[name]; !ok {
		return "", errors.Errorf("no registered annotation with name %s", name)
	}
	if val, found := annotations[annotationRegistry[name].key]; found {
		if err := annotationRegistry[name].validator(val); err != nil {
			log.V(1).Info("error accessing annotation - using default", "name", name)
			return annotationRegistry[name].defaultVal, err
		}
		log.V(1).Info("annotation found", "name", name)
		return val, nil
	}
	log.V(1).Info("annotation not found - using default", "name", name)
	return annotationRegistry[name].defaultVal, nil
}

// NewIntermediateConfig returns a new IntermediateConfig Object constructed from a list of ports and annotations
func NewIntermediateConfig(annotations map[string]string) (*IntermediateConfig, error) {
	intermediateConfig := &IntermediateConfig{}
	valDefaultProbeProxyPort := defaultAppProbeProxyPort

	allFields := map[string]*string{
		"outboundPort":                &intermediateConfig.targetPort,
		"inboundPort":                 &intermediateConfig.inboundPort,
		"ipFamilyMode":                &intermediateConfig.ipFamilyMode,
		"excludeInboundPorts":         &intermediateConfig.excludeInboundPorts,
		"excludeOutboundPorts":        &intermediateConfig.excludeOutboundPorts,
		"isGateway":                   &intermediateConfig.isGateway,
		"builtinDNS":                  &intermediateConfig.builtinDNS,
		"builtinDNSPort":              &intermediateConfig.builtinDNSPort,
		"excludeOutboundPortsForUIDs": &intermediateConfig.excludeOutboundPortsForUIDs,
		"noRedirectUID":               &intermediateConfig.noRedirectUID,
		"applicationProbeProxyPort":   &valDefaultProbeProxyPort,
		"dropInvalidPackets":          &intermediateConfig.dropInvalidPackets,
		"iptablesLogs":                &intermediateConfig.iptablesLogs,
		"excludeInboundIPs":           &intermediateConfig.excludeInboundIPs,
		"excludeOutboundIPs":          &intermediateConfig.excludeOutboundIPs,
	}

	for fieldName, fieldPointer := range allFields {
		if err := mapAnnotation(annotations, fieldPointer, fieldName); err != nil {
			return nil, err
		}
	}

	excludeAppProbeProxyPort(allFields)
	return intermediateConfig, nil
}

func mapAnnotation(annotations map[string]string, field *string, fieldName string) error {
	val, err := getAnnotationOrDefault(fieldName, annotations)
	if err != nil {
		return err
	}
	*field = val
	return nil
}

func excludeAppProbeProxyPort(allFields map[string]*string) {
	inboundPortsToExclude := allFields["excludeInboundPorts"]
	applicationProbeProxyPort := *allFields["applicationProbeProxyPort"]
	if applicationProbeProxyPort == "0" {
		return
	}

	existingExcludes := *inboundPortsToExclude
	if existingExcludes == "" {
		*inboundPortsToExclude = applicationProbeProxyPort
	} else {
		*inboundPortsToExclude = fmt.Sprintf("%s,%s", existingExcludes, applicationProbeProxyPort)
	}
}
