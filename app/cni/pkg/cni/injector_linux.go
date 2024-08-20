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

func convertCommaSeparatedString(list string) (config.Ports, error) {
	split := strings.Split(list, ",")
	mapped := make(config.Ports, len(split))

	for i, value := range split {
		converted, err := convertToUint16(strconv.Itoa(i), value)
		if err != nil {
			return nil, err
		}
		mapped[i] = config.Port(converted)
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

	excludePorts, err := convertCommaSeparatedString(intermediateConfig.excludeOutboundPorts)
	if err != nil {
		return nil, err
	}

	var excludePortsForUIDs []string
	if intermediateConfig.excludeOutboundPortsForUIDs != "" {
		excludePortsForUIDs = strings.Split(intermediateConfig.excludeOutboundPortsForUIDs, ";")
	}

	cfg.CNIMode = true
	cfg.Verbose = true
	cfg.RuntimeStdout = logWriter
	cfg.KumaDPUser.UID = intermediateConfig.noRedirectUID
	cfg.Redirect.Outbound.Enabled = true
	cfg.Redirect.Outbound.ExcludePorts = excludePorts
	cfg.Redirect.Outbound.ExcludePortsForUIDs = excludePortsForUIDs

	if err := cfg.Redirect.Outbound.Port.Set(intermediateConfig.targetPort); err != nil {
		return nil, errors.Wrapf(err, "failed to set outbound port to '%s'", intermediateConfig.targetPort)
	}

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

	if err := cfg.IPFamilyMode.Set(intermediateConfig.ipFamilyMode); err != nil {
		return nil, err
	}

	cfg.Redirect.Inbound.Enabled = !isGateway
	if cfg.Redirect.Inbound.Enabled {
		if err := cfg.Redirect.Inbound.Port.Set(intermediateConfig.inboundPort); err != nil {
			return nil, errors.Wrapf(err, "failed to set inbound port to '%s'", intermediateConfig.inboundPort)
		}

		excludedPorts, err := convertCommaSeparatedString(intermediateConfig.excludeInboundPorts)
		if err != nil {
			return nil, err
		}

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
		if err := cfg.Redirect.DNS.Port.Set(intermediateConfig.builtinDNSPort); err != nil {
			return nil, errors.Wrapf(err, "failed to set builtin dns port to '%s'", intermediateConfig.builtinDNSPort)
		}

		cfg.Redirect.DNS.CaptureAll = true
		cfg.Redirect.DNS.SkipConntrackZoneSplit = false
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
