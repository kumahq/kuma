package cni

import (
	"bufio"
	"bytes"
	"context"
	"strconv"
	"strings"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/transparentproxy"
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
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

func Inject(ctx context.Context, netns string, intermediateConfig *IntermediateConfig, logger logr.Logger) error {
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
		initializedConfig, err := cfg.Initialize(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to initialize config")
		}

		if _, err := transparentproxy.Setup(ctx, initializedConfig); err != nil {
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

func mapToConfig(intermediateConfig *IntermediateConfig, logWriter *bufio.Writer) (*config.Config, error) {
	cfg := config.DefaultConfig()

	port, err := convertToUint16("inbound port", intermediateConfig.targetPort)
	if err != nil {
		return nil, err
	}
	excludePorts, err := convertCommaSeparatedString(intermediateConfig.excludeOutboundPorts)
	if err != nil {
		return nil, err
	}

	var excludePortsForUIDs []string
	if intermediateConfig.excludeOutboundPortsForUIDs != "" {
		excludePortsForUIDs = strings.Split(intermediateConfig.excludeOutboundPortsForUIDs, ";")
	}

	cfg.Verbose = true
	cfg.RuntimeStdout = logWriter
	cfg.Owner.UID = intermediateConfig.noRedirectUID
	cfg.Redirect.Outbound.Enabled = true
	cfg.Redirect.Outbound.Port = port
	cfg.Redirect.Outbound.ExcludePorts = excludePorts
	cfg.Redirect.Outbound.ExcludePortsForUIDs = excludePortsForUIDs

	if intermediateConfig.excludeOutboundIPs != "" {
		cfg.Redirect.Outbound.ExcludePortsForIPs = strings.Split(intermediateConfig.excludeOutboundIPs, ",")
	}

	cfg.DropInvalidPackets, err = GetEnabled(intermediateConfig.dropInvalidPackets)
	if err != nil {
		return nil, err
	}

	cfg.Log.Enabled, err = GetEnabled(intermediateConfig.iptablesLogs)
	if err != nil {
		return nil, err
	}

	isGateway, err := GetEnabled(intermediateConfig.isGateway)
	if err != nil {
		return nil, err
	}

	cfg.IPv6 = intermediateConfig.ipFamilyMode != "ipv4"

	cfg.Redirect.Inbound.Enabled = !isGateway
	if cfg.Redirect.Inbound.Enabled {
		inboundPort, err := convertToUint16("inbound port", intermediateConfig.inboundPort)
		if err != nil {
			return nil, err
		}

		excludedPorts, err := convertCommaSeparatedString(intermediateConfig.excludeInboundPorts)
		if err != nil {
			return nil, err
		}

		cfg.Redirect.Inbound.Port = inboundPort
		cfg.Redirect.Inbound.ExcludePorts = excludedPorts

		if intermediateConfig.excludeInboundIPs != "" {
			cfg.Redirect.Inbound.ExcludePortsForIPs = strings.Split(intermediateConfig.excludeInboundIPs, ",")
		}
	}

	cfg.Redirect.DNS.Enabled, err = GetEnabled(intermediateConfig.builtinDNS)
	if err != nil {
		return nil, err
	}
	if cfg.Redirect.DNS.Enabled {
		builtinDnsPort, err := convertToUint16("builtin dns port", intermediateConfig.builtinDNSPort)
		if err != nil {
			return nil, err
		}

		cfg.Redirect.DNS.Port = builtinDnsPort
		cfg.Redirect.DNS.CaptureAll = true
		cfg.Redirect.DNS.ConntrackZoneSplit = true
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
		return false, errors.Errorf(`wrong value "%s", available values are: "enabled", "disabled", "true", "false"`, value)
	}
}
