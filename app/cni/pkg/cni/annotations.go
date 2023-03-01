package cni

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	defaultProxyStatusPort     = "9901"
	defaultOutboundPort        = "15001"
	defaultInboundPort         = "15006"
	defaultInboundPortV6       = "15010"
	defaultBuiltinDNSPort      = "15053"
	defaultNoRedirectUID       = "5678"
	defaultRedirectExcludePort = defaultProxyStatusPort
)

var annotationRegistry = map[string]*annotationParam{
	"inject":               {"kuma.io/sidecar-injection", "", alwaysValidFunc},
	"ports":                {"kuma.io/envoy-admin-port", "", validatePortList},
	"excludeInboundPorts":  {"traffic.kuma.io/exclude-inbound-ports", defaultRedirectExcludePort, validatePortList},
	"excludeOutboundPorts": {"traffic.kuma.io/exclude-outbound-ports", defaultRedirectExcludePort, validatePortList},
	"inboundPort":          {"kuma.io/transparent-proxying-inbound-port", defaultInboundPort, validatePortList},
	"inboundPortV6":        {"kuma.io/transparent-proxying-inbound-v6-port", defaultInboundPortV6, validatePortList},
	"outboundPort":         {"kuma.io/transparent-proxying-outbound-port", defaultOutboundPort, validatePortList},
	"isGateway":            {"kuma.io/gateway", "false", alwaysValidFunc},
	"builtinDNS":           {"kuma.io/builtindns", "false", alwaysValidFunc},
	"builtinDNSPort":       {"kuma.io/builtindnsport", defaultBuiltinDNSPort, validatePortList},
}

type IntermediateConfig struct {
	targetPort           string
	inboundPort          string
	inboundPortV6        string
	noRedirectUID        string
	excludeInboundPorts  string
	excludeOutboundPorts string
	isGateway            string
	builtinDNS           string
	builtinDNSPort       string
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

func validatePortList(ports string) error {
	if _, err := parsePorts(ports); err != nil {
		return errors.Wrapf(err, "portList %q", ports)
	}
	return nil
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
	intermediateConfig.noRedirectUID = defaultNoRedirectUID

	allFields := map[string]*string{
		"outboundPort":         &intermediateConfig.targetPort,
		"inboundPort":          &intermediateConfig.inboundPort,
		"inboundPortV6":        &intermediateConfig.inboundPortV6,
		"excludeInboundPorts":  &intermediateConfig.excludeInboundPorts,
		"excludeOutboundPorts": &intermediateConfig.excludeOutboundPorts,
		"isGateway":            &intermediateConfig.isGateway,
		"builtinDNS":           &intermediateConfig.builtinDNS,
		"builtinDNSPort":       &intermediateConfig.builtinDNSPort,
	}

	for fieldName, fieldPointer := range allFields {
		if err := mapAnnotation(annotations, fieldPointer, fieldName); err != nil {
			return nil, err
		}
	}

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
