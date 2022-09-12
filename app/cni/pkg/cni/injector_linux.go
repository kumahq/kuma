package cni

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	kumanet_tproxy "github.com/kumahq/kuma-net/transparent-proxy"
	kumanet_config "github.com/kumahq/kuma-net/transparent-proxy/config"

	"github.com/kumahq/kuma/pkg/transparentproxy"
)

func convertToUint16(field string, value string) (uint16, error) {
	converted, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return 0, errors.Wrapf(err, "could not convert field %v", field)
	}
	return uint16(converted), nil
}

func convertCommaSeparatedString(list string) ([]uint16, error) {
	split := strings.Split(list, ",")
	mapped := make([]uint16, len(split))

	for i, value := range split {
		converted, err := convertToUint16(strconv.Itoa(i), value)
		if err != nil {
			return nil, err
		}
		mapped[i] = converted
	}

	return mapped, nil
}

func Inject(netns string, logger logr.Logger, intermediateConfig *IntermediateConfig) error {
	var logBuffer bytes.Buffer
	logWriter := bufio.NewWriter(&logBuffer)
	cfg, err := mapToConfig(intermediateConfig, logWriter)
	if err != nil {
		return err
	}

	namespace, err := ns.GetNS(netns)
	if err != nil {
		return errors.Wrap(err, "failed to open namespace")
	}
	defer namespace.Close()

	return namespace.Do(func(_ ns.NetNS) error {
		if _, err := kumanet_tproxy.Setup(*cfg); err != nil {
			return err
		}

		if err := logWriter.Flush(); err != nil {
			return err
		}

		logger.Info("iptables rules applied")
		logger.V(1).Info("generated iptables rules", "iptablesStdout", logBuffer.String())

		return nil
	})
}

func mapToConfig(intermediateConfig *IntermediateConfig, logWriter *bufio.Writer) (*kumanet_config.Config, error) {
	port, err := convertToUint16("inbound port", intermediateConfig.targetPort)
	if err != nil {
		return nil, err
	}
	excludePorts, err := convertCommaSeparatedString(intermediateConfig.excludeOutboundPorts)
	if err != nil {
		return nil, err
	}
	cfg := kumanet_config.Config{
		RuntimeStdout: logWriter,
		Owner: kumanet_config.Owner{
			UID: intermediateConfig.noRedirectUID,
		},
		Redirect: kumanet_config.Redirect{
			Outbound: kumanet_config.TrafficFlow{
				Enabled:      true,
				Port:         port,
				ExcludePorts: excludePorts,
			},
		},
	}

	isGateway, err := GetEnabled(intermediateConfig.isGateway)
	if err != nil {
		return nil, err
	}
	redirectInbound := !isGateway
	if redirectInbound {
		inboundPort, err := convertToUint16("inbound port", intermediateConfig.inboundPort)
		if err != nil {
			return nil, err
		}
		enableIpV6, err := transparentproxy.ShouldEnableIPv6()
		if err != nil {
			return nil, err
		}
		cfg.IPv6 = enableIpV6

		inboundPortV6, err := convertToUint16("inbound port ipv6", intermediateConfig.inboundPortV6)
		if err != nil {
			return nil, err
		}
		excludedPorts, err := convertCommaSeparatedString(intermediateConfig.excludeInboundPorts)
		if err != nil {
			return nil, err
		}
		cfg.Redirect.Inbound = kumanet_config.TrafficFlow{
			Enabled:      true,
			Port:         inboundPort,
			PortIPv6:     inboundPortV6,
			ExcludePorts: excludedPorts,
		}
	}

	useBuiltinDNS, err := GetEnabled(intermediateConfig.builtinDNS)
	if err != nil {
		return nil, err
	}
	if useBuiltinDNS {
		builtinDnsPort, err := convertToUint16("builtin dns port", intermediateConfig.builtinDNSPort)
		if err != nil {
			return nil, err
		}
		cfg.Redirect.DNS = kumanet_config.DNS{
			Enabled:            true,
			Port:               builtinDnsPort,
			CaptureAll:         true,
			ConntrackZoneSplit: true,
		}
	}
	return &cfg, nil
}

func GetEnabled(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "enabled", "true":
		return true, nil
	case "disabled", "false":
		return false, nil
	default:
		return false, errors.Errorf(`wrong value "%s", available values are: "enabled", "disabled"`, value)
	}
}
